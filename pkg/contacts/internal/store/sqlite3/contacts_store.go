package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/mattn/go-sqlite3"
)

type ContactsStore struct {
	db *sql.DB
}

func NewContactsStore(saveDir string, noSave bool) (store *ContactsStore, isNew bool, err error) {
	db, isNew, err := openDB(saveDir, noSave)
	if err != nil {
		return nil, false, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60 * 1e9) // 5 minutes in nanoseconds

	if err = initDB(db); err != nil {
		return
	}
	store = &ContactsStore{db: db}

	return store, isNew, nil
}

func (store *ContactsStore) Begin() (*sql.Tx, error) {
	return store.db.Begin()
}

func (store *ContactsStore) GetGroup(tx *sql.Tx, groupId int) (*model.Group, error) {
	var group model.Group
	err := tx.QueryRow(`
	SELECT id, name, created_at
	FROM groups
	WHERE id = ?
	`, groupId).Scan(&group.Id, &group.Name, &group.CreatedAt)
	if err != nil {
		return nil, wrapErr(err)
	}
	return &group, nil
}

// ListGroups returns a list of groups
func (store *ContactsStore) ListGroups(tx *sql.Tx, limit int) ([]model.Group, error) {
	// TODO: add query param to search by name, time
	rows, err := tx.Query(`
	SELECT id, name, created_at
	FROM groups
	LIMIT ?;
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]model.Group, 0)
	for rows.Next() {
		var group model.Group
		if err = rows.Scan(&group.Id, &group.Name, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (store *ContactsStore) ListUserGroups(tx *sql.Tx, userId int32) ([]model.Group, error) {
	exists, err := store.profileExists(tx, userId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrUserNotExists
	}

	groups, err := store.listUserGroups(tx, userId)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (store *ContactsStore) listUserGroups(tx *sql.Tx, userId int32) ([]model.Group, error) {
	rows, err := tx.Query(`
	SELECT g.id, g.name, g.created_at
	FROM groups g
	JOIN member m ON g.id = m.group_id
	WHERE m.user_id = ?`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]model.Group, 0)
	for rows.Next() {
		var group model.Group
		if err = rows.Scan(&group.Id, &group.Name, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (store *ContactsStore) CreateGroup(tx *sql.Tx, name string) (*model.Group, error) {
	if len(name) < 2 {
		return nil, fmt.Errorf("group name must be at least 2 characters long")
	}
	row := tx.QueryRow(`
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

func (store *ContactsStore) DeleteGroup(tx *sql.Tx, id int) error {
	result, err := tx.Exec(`
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

// MemberExists checks member exists in group
func (store *ContactsStore) MemberExists(tx *sql.Tx, groupId int, userId int32) (bool, error) {
	var memberExists bool

	var profileExists, groupExists bool
	err := tx.QueryRow(`
	SELECT
		EXISTS(SELECT 1 FROM profile WHERE user_id = ?) AS profile_exists,
		EXISTS(SELECT 1 FROM groups WHERE id = ?) AS group_exists,
		EXISTS(SELECT 1 FROM member WHERE group_id = ? AND user_id = ?) as member_exists;
	`, userId, groupId, groupId, userId).Scan(&profileExists, &groupExists, &memberExists)
	if err != nil {
		return false, err
	}

	if !profileExists {
		return false, errs.ErrUserNotExists
	}
	if !groupExists {
		return false, errs.ErrGroupNotExists
	}
	return memberExists, nil
}

func (store *ContactsStore) CreateMember(tx *sql.Tx, groupId int, userId int32, role access.Role) (*model.Member, error) {
	exists, err := store.MemberExists(tx, groupId, userId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrExists
	}

	row := tx.QueryRow(`
	INSERT INTO member (group_id, user_id, role)
	VALUES (?, ?, ?)
	RETURNING group_id, user_id, created_at, role`, groupId, userId, role)
	var member model.Member
	if err = row.Scan(&member.GroupId, &member.UserId, &member.CreatedAt, &member.Role); err != nil {
		return nil, err
	}
	return &member, nil
}

func (store *ContactsStore) GetMember(tx *sql.Tx, groupId int, userId int32) (*model.Member, error) {
	var member model.Member
	err := tx.QueryRow(`
	SELECT group_id, user_id, created_at, role
	FROM member
	WHERE 
		group_id = ? AND
		user_id = ?;
	`, groupId, userId).Scan(&member.GroupId, &member.UserId, &member.CreatedAt, &member.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (store *ContactsStore) ListGroupMembers(tx *sql.Tx, groupId int) ([]model.Member, error) {
	rows, err := tx.Query(`
	SELECT group_id, user_id, created_at, role
	FROM member
	WHERE group_id = ?;
	`, groupId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.Member, 0)
	for rows.Next() {
		var member model.Member
		if err = rows.Scan(&member.GroupId, &member.UserId, &member.CreatedAt, &member.Role); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (store *ContactsStore) DeleteMember(tx *sql.Tx, groupId int, userId int32) error {
	exists, err := store.MemberExists(tx, groupId, userId)
	if err != nil {
		return err
	}
	if !exists {
		return errs.ErrNotFound
	}

	result, err := tx.Exec(`
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

func (store *ContactsStore) profileExists(tx *sql.Tx, id int32) (bool, error) {
	var profileExists bool
	err := tx.QueryRow(`
	SELECT 
		EXISTS (SELECT 1 FROM profile WHERE user_id = ?);
	`, id).Scan(&profileExists)
	if err != nil {
		return false, err
	}
	return profileExists, nil
}

func (store *ContactsStore) ListProfiles(tx *sql.Tx, limit int) ([]model.Profile, error) {
	rows, err := tx.Query(`
	SELECT user_id, name, icon_url
	FROM profile
	LIMIT ?;
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]model.Profile, 0)
	for rows.Next() {
		var profile model.Profile
		if err = rows.Scan(&profile.Id, &profile.Name); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (store *ContactsStore) CreateProfile(tx *sql.Tx, userId int32, name string) (*model.Profile, error) {
	exists, err := store.profileExists(tx, userId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errs.ErrExists
	}

	row := tx.QueryRow(`
	INSERT INTO profile (user_id, name)
	VALUES (?, ?)
	RETURNING user_id, name;
	`, userId, name)
	var profile model.Profile

	err = row.Scan(&profile.Id, &profile.Name)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (store *ContactsStore) UpdateProfile(tx *sql.Tx, userId int32, name, iconUrl string) (*model.Profile, error) {
	if exists, err := store.profileExists(tx, userId); !exists {
		if err != nil {
			return nil, err
		}
		return nil, errs.ErrUserNotExists
	}

	_, err := tx.Exec(`
	UPDATE profile
	SET name = ?, icon_url = ?
	WHERE user_id = ?;
	`, name, iconUrl, userId)
	if err != nil {
		return nil, err
	}

	return store.GetProfile(tx, userId)
}

func (store *ContactsStore) DeleteProfile(tx *sql.Tx, id int32) error {
	if exists, err := store.profileExists(tx, id); !exists {
		if err != nil {
			return err
		}
		return errs.ErrUserNotExists
	}

	result, err := tx.Exec(`
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

func (store *ContactsStore) GetProfile(tx *sql.Tx, userId int32) (*model.Profile, error) {
	var profile model.Profile
	err := tx.QueryRow(`
	SELECT user_id, name, icon_url
	FROM profile
	WHERE user_id = ?
	`, userId).Scan(&profile.Id, &profile.Name)
	if err != nil {
		return nil, wrapErr(err)
	}
	return &profile, nil
}

func (store *ContactsStore) CreateContact(tx *sql.Tx, ownerId, targetId int32, alias *string) (*model.Contact, error) {
	// owner exists
	if exists, err := store.profileExists(tx, ownerId); !exists {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("owner %d: %w", ownerId, errs.ErrUserNotExists)
	}

	// target exists
	if exists, err := store.profileExists(tx, targetId); !exists {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("target %d: %w", targetId, errs.ErrUserNotExists)
	}

	if alias != nil && *alias == "" {
		alias = nil
	}

	// TODO: improve query
	row := tx.QueryRow(`
	INSERT INTO contacts (owner_id, target_id, alias)
	VALUES (?, ?, ?)
	RETURNING target_id, (SELECT name FROM profile WHERE user_id = target_id), alias, created_at;
	`, ownerId, targetId, alias)

	var (
		contact model.Contact
		dbAlias sql.NullString
	)
	if err := row.Scan(&contact.ProfileId, &contact.Name, &dbAlias, &contact.CreatedAt); err != nil {
		return nil, err
	}
	contact.Alias = nullStringPtr(dbAlias)
	return &contact, nil
}

func (store *ContactsStore) ListUserContacts(tx *sql.Tx, userId int32) ([]model.Contact, error) {
	if exists, err := store.profileExists(tx, userId); !exists {
		if err != nil {
			return nil, err
		}
		return nil, errs.ErrUserNotExists
	}

	rows, err := tx.Query(`
	SELECT p.user_id, p.name, c.alias, c.created_at
	FROM contacts c
	JOIN profile p ON c.target_id = p.user_id
	WHERE c.owner_id = ?;
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contacts := make([]model.Contact, 0)
	for rows.Next() {
		var (
			contact model.Contact
			alias   sql.NullString
		)
		if err = rows.Scan(&contact.ProfileId, &contact.Name, &alias, &contact.CreatedAt); err != nil {
			return nil, err
		}
		contact.Alias = nullStringPtr(alias)
		contacts = append(contacts, contact)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return contacts, nil
}

func (store *ContactsStore) DeleteUserContact(tx *sql.Tx, ownerId, targetId int32) error {
	if exists, err := store.profileExists(tx, ownerId); !exists {
		if err != nil {
			return err
		}
		return errs.ErrUserNotExists
	}

	result, err := tx.Exec(`
	DELETE FROM contacts
	WHERE owner_id = ? AND target_id = ?;
	`, ownerId, targetId)
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

func (store *ContactsStore) FindOwners(tx *sql.Tx, userId int32) ([]model.Member, error) {
	rows, err := tx.Query(`
	SELECT group_id, user_id, created_at, role
	FROM member
	WHERE user_id = ? AND role = ?;
	`, userId, access.RoleOwner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.Member, 0)
	for rows.Next() {
		var member model.Member
		if err = rows.Scan(
			&member.GroupId,
			&member.UserId,
			&member.CreatedAt,
			&member.Role,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// wrapErr converts sqlite3.Error to store.Error
// if the error cannot be converted, it returns the original error
func wrapErr(err error) error {
	var sqlErr sqlite3.Error
	errors.As(err, &sqlErr)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}
	return err
}

func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func openDB(saveDir string, inMemory bool) (db *sql.DB, isNew bool, err error) {
	if inMemory {
		if db, err = sql.Open("sqlite3", "file:contacts_memdb?mode=memory&cache=shared"); err != nil {
			return nil, false, err
		}
		return db, true, nil
	}
	path := filepath.Join(saveDir, "contacts.db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		isNew = true
	}
	if db, err = sql.Open("sqlite3", path); err != nil {
		return nil, false, err
	}
	return db, isNew, nil
}

// initDB initializes the database schema
func initDB(db *sql.DB) (err error) {
	errs := make([]error, 0)

	// Enable foreign key constraints
	_, err = db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		errs = append(errs, fmt.Errorf("enable foreign keys: %w", err))
	}

	tableQueries := map[string]string{
		"create table profile": `
	CREATE TABLE IF NOT EXISTS profile (
	user_id INTEGER PRIMARY KEY,
	name varchar(20) NOT NULL
	);`,
		"create table groups": `
	CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name varchar(50) NOT NULL CHECK(length(name) >= 2),
	created_at TIMESTAMP DEFAULT (datetime('now'))
	);`,
		"create table member": `
	CREATE TABLE IF NOT EXISTS member (
	user_id INTEGER NOT NULL,
	group_id INTEGER NOT NULL,
	role INTEGER NOT NULL,
	created_at TIMESTAMP DEFAULT (datetime('now')),
	FOREIGN KEY(user_id) REFERENCES profile(user_id) ON DELETE CASCADE,
	FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
	PRIMARY KEY(user_id, group_id)
	);`,
		"create table contacts": `
	CREATE TABLE IF NOT EXISTS contacts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	owner_id INTEGER NOT NULL,
	target_id INTEGER NOT NULL,
	alias varchar(50),
	created_at TIMESTAMP DEFAULT (datetime('now')),
	FOREIGN KEY(owner_id) REFERENCES profile(user_id) ON DELETE CASCADE,
	FOREIGN KEY(target_id) REFERENCES profile(user_id) ON DELETE CASCADE
	);`,
	}

	indexQueries := map[string]string{
		"create index idx_contacts_owner": `
	CREATE INDEX IF NOT EXISTS idx_contacts_owner
	ON contacts (owner_id);`,
		"create index idx_member_user": `
	CREATE INDEX IF NOT EXISTS idx_member_user
	ON member (user_id);`,
	}

	for name, query := range tableQueries {
		_, err = db.Exec(query)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}

	for name, query := range indexQueries {
		_, err = db.Exec(query)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}
