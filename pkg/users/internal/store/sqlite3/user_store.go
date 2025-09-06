package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elug3/gochat/pkg/users/internal/errs"
	"github.com/elug3/gochat/pkg/users/internal/model"
	"github.com/elug3/gochat/pkg/users/internal/store"
	_ "github.com/mattn/go-sqlite3"
)

type UserStore struct {
	db *sql.DB
}

type TxUser struct {
	tx *sql.Tx
}

func NewUserStore() (*UserStore, error) {
	db, err := openDB("")
	if err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	if err = initDB(db); err != nil {
		return nil, err
	}

	store := UserStore{db: db}
	return &store, nil
}

func (store *UserStore) BeginTx(ctx context.Context) (store.TxUser, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TxUser{tx: tx}, nil
}

func (store *UserStore) Begin() (store.TxUser, error) {
	tx, err := store.db.Begin()
	if err != nil {
		return nil, err
	}
	return &TxUser{tx: tx}, nil
}

func (txu *TxUser) Rollback() error {
	return txu.tx.Rollback()
}

func (txu *TxUser) Commit() error {
	return txu.tx.Commit()
}

func (txu *TxUser) userExists(username string) (bool, error) {
	var userExists bool
	err := txu.tx.QueryRow(`
	SELECT EXISTS (SELECT 1 FROM users WHERE username = ?) as user_exists;
	`, username).Scan(&userExists)
	if err != nil {
		return false, err
	}
	return userExists, nil
}

func (txu *TxUser) CreateUser(username string) (*model.User, error) {
	if len(username) < 2 {
		return nil, errs.ErrInvalid
	}

	if exists, _ := txu.userExists(username); exists {
		return nil, errs.ErrExists
	}
	return txu.createUser(username)
}

func (txu *TxUser) createUser(username string) (*model.User, error) {
	var user model.User
	err := txu.tx.QueryRow(`
	INSERT INTO users (username)
	VALUES (?)
	RETURNING id, username;
	`, username).Scan(&user.Id, &user.Username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (txu *TxUser) GetUser(userId int32) (*model.User, error) {
	row := txu.tx.QueryRow(`
	SELECT id, username 
	FROM users 
	WHERE id = ?;
	`, int(userId))
	var user model.User
	if err := row.Scan(&user.Id, &user.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (txu *TxUser) UpdateUser(userId int32, username string) (*model.User, error) {
	if len(username) < 2 {
		return nil, errs.ErrInvalid
	}
	_, err := txu.tx.Exec(`
	UPDATE users SET username = ? WHERE id = ?;
	`, username, int(userId))
	if err != nil {
		return nil, err
	}
	return txu.GetUser(userId)
}

func (txu *TxUser) DeleteUser(userId int32) error {
	res, err := txu.tx.Exec(`
	DELETE FROM users WHERE id = ?;
	`, int(userId))
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func openDB(saveDir string) (*sql.DB, error) {
	if saveDir == "" {
		return sql.Open("sqlite3", ":memory:")
	}
	dbPath := saveDir + "/users.db"
	return sql.Open("sqlite3", dbPath)
}

func initDB(db *sql.DB) error {
	errsSlice := make([]error, 0)
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username VARCHAR(20) NOT NULL UNIQUE CHECK(length(username) >= 2) 
	);`)
	if err != nil {
		errsSlice = append(errsSlice, fmt.Errorf("create table user: %w", err))
	}
	// Removed password and access_token tables creation as authentication methods are removed.
	return errors.Join(errsSlice...)
}
