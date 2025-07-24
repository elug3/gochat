package sqlite

import (
	"crypto/rand"
	"crypto/sha3"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
)

type UserStore struct {
	db *sql.DB
}

type TxUser struct {
	tx *sql.Tx
}

func NewUserStore(cfg *config.Config) (*UserStore, error) {
	db, err := openDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	if err = initDB(db); err != nil {
		return nil, err
	}

	store := UserStore{db: db}
	return &store, nil
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
		return nil, &store.Error{
			Kind:    store.KindUser,
			Err:     store.ErrBadRequest,
			Message: "username must be at least 2 characters long",
		}
	}

	if exists, _ := txu.userExists(username); exists {
		return nil, &store.Error{
			Kind:    store.KindUser,
			Err:     store.ErrExists,
			Message: fmt.Sprintf("username '%s' already exists", username),
		}
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

func (txu *TxUser) GetUser(userId int) (*model.User, error) {
	row := txu.tx.QueryRow(`
	SELECT id, username 
	FROM users 
	WHERE id = ?
	;`, userId)
	var user model.User
	if err := row.Scan(&user.Id, &user.Username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &store.Error{
				Kind:    store.KindUser,
				Err:     store.ErrNotFound,
				Message: fmt.Sprintf("user '%d' not found", userId),
			}
		}
		return nil, err
	}
	return &user, nil
}

func (txu *TxUser) UpdatePassword(userId int, password string) error {
	newHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	_, err = txu.tx.Exec(`
	INSERT INTO password (user_id, hash, created_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id) DO UPDATE SET
		hash = excluded.hash,
		created_at = CURRENT_TIMESTAMP;
	`, userId, newHash)
	return err
}

func (txu *TxUser) ValidatePassword(username, password string) (userId int, err error) {
	row := txu.tx.QueryRow(`
	SELECT u.id, p.hash
	FROM password p
	JOIN users u ON u.id = p.user_id
	WHERE u.username = ?;
	`, username)

	var passwordHash string
	if err = row.Scan(&userId, &passwordHash); err != nil {
		return 0, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, passwordHash)
	if err != nil {
		return 0, err
	}
	if !match {
		return 0, &store.Error{
			Kind:    store.KindUser,
			Err:     store.ErrBadRequest,
			Message: "",
		}
	}
	return userId, nil
}

func generateId(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(b), nil
}

func hashText(s string) string {
	sum := sha3.Sum256([]byte(s))
	return base32.HexEncoding.EncodeToString(sum[:])
}

func (txu *TxUser) CreateToken(userId int, expiresIn time.Duration) (*model.Token, error) {
	tokenString := rand.Text()
	hash := hashText(tokenString)

	var expiresAt sql.NullTime
	if expiresIn != 0 {
		expiresAt = sql.NullTime{Time: time.Now().Add(expiresIn), Valid: true}
	} else {
		expiresAt = sql.NullTime{Valid: false}
	}

	row := txu.tx.QueryRow(`
	INSERT INTO access_token (user_id, hash, expires_at)
	VALUES (?, ?, ?)
	RETURNING id, user_id, expires_at, issued_at;
	`, userId, hash, expiresAt)
	var token model.Token
	if err := row.Scan(&token.Id, &token.UserId, &token.ExpiresAt, &token.IssuedAt); err != nil {
		return nil, err
	}
	token.AccessToken = tokenString
	return &token, nil
}

func (txu *TxUser) GetToken(tokenString string) (*model.Token, error) {
	var token model.Token
	hash := hashText(tokenString)
	err := txu.tx.QueryRow(`
	SELECT id, user_id, expires_at, issued_at
	FROM access_token 
	WHERE hash = ?;`, hash).Scan(&token.Id, &token.UserId, &token.ExpiresAt, &token.IssuedAt)
	if err != nil {
		return nil, err
	}
	return &token, nil

}

func (txu *TxUser) ListTokens(userId int) ([]model.Token, error) {
	rows, err := txu.tx.Query(`
	SELECT id, user_id, expires_at, issued_at
	FROM access_token
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]model.Token, 0)
	for rows.Next() {
		var token model.Token
		if err := rows.Scan(&token.Id, &token.UserId, &token.ExpiresAt, &token.IssuedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

// splitEdge returns the first and last 4 characters of the given string.
func splitEdge(s string) (first, late string) {
	return s[:4], s[len(s)-4:]
}

func openDB(cfg *config.Config) (*sql.DB, error) {
	var path string
	if cfg.NoSave {
		path = ":memory:"
	} else {
		path = cfg.SaveDir + "/contacts.db"
	}
	return sql.Open("sqlite3", path)
}

func initDB(db *sql.DB) error {
	errs := make([]error, 0)
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username VARCHAR(20) NOT NULL UNIQUE CHECK(length(username) >= 2) 
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table user: %w", err))
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS password (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL UNIQUE,
	hash TEXT NOT NULL,
	created_at DATETIME DEFAULT (datetime('now')),
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table password: %w", err))
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS access_token (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL UNIQUE,
	hash TEXT NOT NULL,
	expires_at TIMESTAMP,
	issued_at TIMESTAMP DEFAULT (datetime('now')),
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("create table acces_tokens: %w", err))
	}
	return errors.Join(errs...)
}
