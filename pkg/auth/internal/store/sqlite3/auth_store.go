package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/mattn/go-sqlite3"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore() (*AuthStore, error) {
	db, err := sql.Open("sqlite3", "./auth.db")
	if err != nil {
		return nil, err
	}
	if err = initdb(db); err != nil {
		db.Close()
		return nil, err
	}

	return &AuthStore{db: db}, nil
}

func (store *AuthStore) GetCredential(ctx context.Context, username string) (*model.Credentials, error) {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	var c model.Credentials
	err = tx.QueryRow(`
	SELECT user_id, username, password_hash, updated_at
	FROM credentials
	WHERE username = ?;`, username).Scan(&c.UserId, &c.Username, &c.PasswordHash, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (store *AuthStore) CreateCredential(ctx context.Context, username string, passwordHash string) (int32, error) {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`
	INSERT INTO credentials (username, password_hash)
	VALUES (?, ?);
	`, username, passwordHash)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return 0, errs.ErrExists
			}
			return 0, err
		}
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return int32(userId), nil
}

func (store *AuthStore) SaveCredentials(ctx context.Context, userId int32, username string, passwordHash string) error {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	INSERT INTO credentials (user_id, username, password_hash)
	VALUES (?, ?, ?);
	`, userId, username, passwordHash)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return errs.ErrExists
			}
			return err
		}
	}

	return tx.Commit()
}

func (store *AuthStore) UpdatePassword(ctx context.Context, userId int32, newPasswordHash string) error {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
	UPDATE credentials
	SET password_hash = ? 
	WHERE user_id = ?;
	`, newPasswordHash, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}
		return err
	}

	return tx.Commit()
}

func (store *AuthStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func initdb(db *sql.DB) error {
	var errs []error
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS credentials (
	user_id INTEGER PRIMARY KEY,
	username TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot create table credentials: %w", err))
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS refresh_tokens (
	token_id    UUID PRIMARY KEY,
	user_id     UUID NOT NULL REFERENCES credentials(user_id) ON DELETE CASCADE,
	token       VARCHAR(512) NOT NULL,
	expires_at  TIMESTAMP NOT NULL,
	created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	revoked     BOOLEAN NOT NULL DEFAULT FALSE
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot create table refresh_tokens: %w", err))
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot create index idx_refresh_tokens_user: %w", err))
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot create index idx_refresh_tokens_token: %w", err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS signing_keys (
	key_id     UUID PRIMARY KEY,
	public_key TEXT NOT NULL,
	private_key TEXT NOT NULL,
	algorithm  VARCHAR(50) NOT NULL,
	active     BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot create table signing_keys: %w", err))
	}

	return errors.Join(errs...)
}
