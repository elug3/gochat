package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
	_ "github.com/mattn/go-sqlite3"
)

type ContactsStore struct {
	db *sql.DB
}

type TxContacts struct {
	tx *sql.Tx
}

func NewContactsStore(cfg *config.Config) (*ContactsStore, error) {
	db, err := openDB(cfg)
	if err != nil {
		return nil, err
	}
	if err = initDB(db); err != nil {
		return nil, err
	}
	store := ContactsStore{db: db}
	return &store, nil
}

func (store *ContactsStore) Begin() (store.TxContacts, error) {
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}
	return &TxContacts{tx: tx}, nil
}

func (txc *TxContacts) Rollback() error {
	return txc.tx.Rollback()
}

func (txc *TxContacts) Commit() error {
	return txc.tx.Commit()
}

func (txc *TxContacts) GetGroup(groupId int) (*model.Group, error) {
	var group model.Group
	err := txc.tx.QueryRow(`
	SELECT id, name, created_at
	FROM groups
	WHERE id = ?
	`, groupId).Scan(&group.Id, &group.Name, &group.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (txc *TxContacts) GetGroups(userId int) ([]model.Group, error) {
	exists, err := txc.profileExists(userId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &store.Error{
			Kind:    store.KindProfile,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("profile '%d' not found.", userId),
		}
	}

	if userId <= 0 {
		return nil, &store.Error{
			Kind:    store.KindUser,
			Err:     store.ErrBadRequest,
			Message: "userId must be greater than 0.",
		}
	}
	groups, err := txc.getGroups(userId)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (txc *TxContacts) getGroups(userId int) ([]model.Group, error) {
	rows, err := txc.tx.Query(`
	SELECT g.id, g.name, g.created_at
	FROM groups g
	JOIN member m ON g.id = m.group_id
	WHERE m.user_id = ?`, userId)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	groups := make([]model.Group, 0)
	for rows.Next() {
		var group model.Group
		if err = rows.Scan(&group.Id, &group.Name, &group.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return groups, nil
}

func (txc *TxContacts) CreateGroup(name string) (*model.Group, error) {
	if len(name) < 2 {
		return nil, &store.Error{
			Kind:    store.KindGroup,
			Err:     store.ErrBadRequest,
			Message: "Group name must be at least two characters long",
		}
	}
	row := txc.tx.QueryRow(`
	INSERT INTO groups (name)
	VALUES (?)
	RETURNING id, name, created_at`, name)
	var group model.Group
	err := row.Scan(&group.Id, &group.Name, &group.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (txc *TxContacts) DeleteGroup(id int) error {
	result, err := txc.tx.Exec(`
	DELETE FROM groups
	WHERE id = ?;
	`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected to delete 1 row, but deleted %d", n)
	}
	return nil
}

// memberExists checks member exists in group
func (txc *TxContacts) MemberExists(groupId, userId int) (bool, error) {
	var memberExists bool

	var profileExists, groupExists bool
	err := txc.tx.QueryRow(`
	SELECT
		EXISTS(SELECT 1 FROM profile WHERE user_id = ?) AS profile_exists,
		EXISTS(SELECT 1 FROM groups WHERE id = ?) AS group_exists,
		EXISTS(SELECT 1 FROM member WHERE group_id = ? AND user_id = ?) as member_exists;
	`, userId, groupId, groupId, userId).Scan(&profileExists, &groupExists, &memberExists)
	if err != nil {
		return false, fmt.Errorf("query.checkMemberExists: %w", err)
	}

	if !profileExists {
		return false, &store.Error{
			Kind:    store.KindProfile,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("user or profile %q not found", userId),
		}
	}
	if !groupExists {
		return false, &store.Error{
			Kind:    store.KindGroup,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("group %q not found", groupId),
		}
	}
	return memberExists, nil
}

func (txc *TxContacts) CreateMember(groupId, userId int, role model.Role) (*model.Member, error) {
	exists, err := txc.MemberExists(groupId, userId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &store.Error{
			Kind:    store.KindGroup,
			Err:     store.ErrExists,
			Message: fmt.Sprintf("user %q already exists in group %q", userId, groupId),
		}
	}

	row := txc.tx.QueryRow(`
	INSERT INTO member (group_id, user_id, role)
	VALUES (?, ?, ?)
	RETURNING group_id, user_id, created_at, role`, groupId, userId, role)
	var member model.Member
	if err = row.Scan(&member.GroupId, &member.UserId, &member.CreatedAt, &member.Role); err != nil {
		return nil, err
	}
	return &member, nil
}

func (txc *TxContacts) GetMember(groupId, userId int) (*model.Member, error) {
	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, err
		}
		return nil, &store.Error{
			Kind:    store.KindMember,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("user '%d' not exists in group '%d'", userId, groupId),
		}
	}

	var member model.Member
	err := txc.tx.QueryRow(`
	SELECT group_id, user_id, created_at, role
	FROM member
	WHERE 
		group_id = ? AND
		user_id = ?;
	`, groupId, userId).Scan(&member.GroupId, &member.UserId, &member.CreatedAt, &member.Role)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (txc *TxContacts) GetMembers(groupId int) ([]model.Member, error) {
	rows, err := txc.tx.Query(`
	SELECT group_id, user_id, role, created_at
	FROM member
	WHERE group_id = ?
	`, groupId)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	members := make([]model.Member, 0)
	for rows.Next() {
		var m model.Member
		err = rows.Scan(
			&m.GroupId,
			&m.UserId,
			&m.Role,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		members = append(members, m)
	}
	return members, nil

}

func (txc *TxContacts) DeleteMember(groupId, userId int) error {
	exists, err := txc.MemberExists(groupId, userId)
	if err != nil {
		return err
	}
	if !exists {
		return &store.Error{
			Kind:    store.KindGroup,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("member %q not found in group %q", userId, groupId),
		}
	}

	result, err := txc.tx.Exec(`
	DELETE FROM member
	WHERE group_id = ? AND user_id = ?;
	`, groupId, userId)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected to delete 1 row, but deleted %d", n)
	}
	return nil
}

func (txc *TxContacts) profileExists(id int) (bool, error) {
	var profileExists bool
	err := txc.tx.QueryRow(`
	SELECT 
		EXISTS (SELECT 1 FROM profile WHERE user_id = ?);
	`, id).Scan(&profileExists)
	if err != nil {
		return false, err
	}
	return profileExists, nil
}

func (txc *TxContacts) CreateProfile(userId int, name string) (*model.Profile, error) {
	if exists, _ := txc.profileExists(userId); exists {
		return nil, &store.Error{
			Kind:    store.KindProfile,
			Err:     store.ErrExists,
			Message: fmt.Sprintf("profile %q already exists", userId),
		}
	}

	row := txc.tx.QueryRow(`
	INSERT INTO profile (user_id, name)
	VALUES (?, ?)
	RETURNING user_id, name;
	`, userId, name)
	var profile model.Profile

	err := row.Scan(&profile.Id, &profile.Name)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (txc *TxContacts) DeleteProfile(id int) error {
	if exists, err := txc.profileExists(id); !exists {
		if err != nil {
			return fmt.Errorf("profileExists: %w", err)
		}
		return &store.Error{
			Kind:    store.KindProfile,
			Err:     store.ErrNotFound,
			Message: fmt.Sprintf("profile %q not found", id),
		}
	}

	result, err := txc.tx.Exec(`
	DELETE FROM profile
	WHERE user_id = ?
	`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected 1 row affected, but got: %d", n)
	}
	return nil
}

func openDB(cfg *config.Config) (*sql.DB, error) {
	var path string
	if cfg.NoSave {
		path = ":memory:"
	} else {
		path = "file:" + cfg.SaveDir + "/contacts.db"
	}
	return sql.Open("sqlite3", path)
}

func initDB(db *sql.DB) error {
	errs := make([]error, 0)

	// _, err := db.Exec(`PRAGMA foreign_keys = ON;`)
	// if err != nil {
	// 	errs = append(errs, err)
	// }

	_, err := db.Exec(`
	 
	CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name varchar(50) NOT NULL CHECK(length(name) >= 2),
	created_at TIMESTAMP DEFAULT (datetime('now'))
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table group: %w", err))
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS member (
	user_id INTEGER NOT NULL,
	group_id INTEGER NOT NULL,
	role INTEGER NOT NULL,
	created_at TIMESTAMP DEFAULT (datetime('now')),
	FOREIGN KEY(user_id) REFERENCES profile(user_id) ON DELETE CASCADE,
	FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
	PRIMARY KEY(user_id, group_id)
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table member: %w", err))
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS profile (
	user_id INTEGER PRIMARY KEY,
	name varchar(20) NOT NULL
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table profile: %w", err))
	}
	return errors.Join(errs...)
}
