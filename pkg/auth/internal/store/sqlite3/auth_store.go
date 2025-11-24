package sqlite3

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/mattn/go-sqlite3"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(saveDir string, inMemory bool) (*AuthStore, error) {
	var db *sql.DB
	var path string
	if inMemory {
		path = ":memory:"
	} else {
		path = saveDir + "/auth.db"
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err = initdb(ctx, db); err != nil {
		return nil, fmt.Errorf("cannot initialize database: %w", err)
	}

	return &AuthStore{db: db}, nil
}

func (store *AuthStore) DB() *sql.DB {
	if store.db == nil {
		panic("database not initialized")
	}
	return store.db
}

func (store *AuthStore) CreateUser(ctx context.Context, tx *sql.Tx, username string) (*model.User, error) {
	result, err := tx.QueryContext(ctx, `
	INSERT INTO users (username)
	VALUES (?)
	RETURNING id, username;
	`, username)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user: %w", err)
	}
	defer result.Close()

	var u model.User
	if result.Next() {
		if err := result.Scan(&u.Id, &u.Username); err != nil {
			return nil, fmt.Errorf("cannot scan inserted user: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no user returned after insert")
	}

	return &u, nil
}

func (store *AuthStore) GetUserById(ctx context.Context, tx *sql.Tx, userId int32) (*model.User, error) {
	var u model.User
	err := tx.QueryRowContext(ctx, `
	SELECT id, username
	FROM users
	WHERE id = ?;`, userId).Scan(&u.Id, &u.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
	}
	return &u, nil
}

func (store *AuthStore) SetPasswordHash(ctx context.Context, tx *sql.Tx, userId int32, passwordHash string) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO passwords (user_id, password_hash)
	VALUES (?, ?)
	ON CONFLICT(user_id) DO UPDATE SET password_hash=excluded.password_hash, updated_at=CURRENT_TIMESTAMP;
	`, userId, passwordHash)
	if err != nil {
		return fmt.Errorf("cannot update password hash: %w", err)
	}

	return nil
}

func (store *AuthStore) GetPasswordByUsername(ctx context.Context, tx *sql.Tx, username string) (*model.PasswordCredential, error) {
	var pw model.PasswordCredential
	err := tx.QueryRowContext(ctx, `
	SELECT p.user_id, u.username, p.password_hash
	FROM passwords as p
	JOIN users as u ON u.id = p.user_id
	WHERE u.username = ?;`, username).Scan(&pw.UserId, &pw.Username, &pw.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("cannot get password: %w", err)
	}
	return &pw, nil
}

func (store *AuthStore) SaveSessionData(ctx context.Context, tx *sql.Tx, session *webauthn.SessionData) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("cannot marshal session data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO webauthn_sessions (challenge, session_data)
	VALUES (?, ?);
	`, session.Challenge, data)
	if err != nil {
		return fmt.Errorf("cannot insert session data: %w", err)
	}

	return nil
}

func (store *AuthStore) GetSessionData(ctx context.Context, tx *sql.Tx, challenge string) (*webauthn.SessionData, error) {
	row := tx.QueryRowContext(ctx, `
	SELECT session_data
	FROM webauthn_sessions
	WHERE challenge = ?;
	`, challenge)

	var data []byte
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("cannot scan session data: %w", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("cannot unmarshal session data: %w", err)
	}

	return &session, nil
}

func (store *AuthStore) DeleteSessionData(ctx context.Context, tx *sql.Tx, challenge string) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM webauthn_sessions
	WHERE challenge = ?;
	`, challenge)
	if err != nil {
		return fmt.Errorf("cannot delete session data: %w", err)
	}

	return nil
}

func (store *AuthStore) SaveWebAuthnCredential(ctx context.Context, tx *sql.Tx, userId int32, credential *webauthn.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("cannot marshal credential data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO webauthn_credentials (user_id, credential_data)
	VALUES (?, ?);
	`, userId, data)
	if err != nil {
		return fmt.Errorf("cannot insert credential data: %w", err)
	}

	return nil
}

// GetWebAuthnUser returns the WebAuthnUser for the given userId.
// It returns empty credentials user for valid user with no credentials.
// If user does not exist, returns errs.ErrNotFound.
func (store *AuthStore) GetWebAuthnUser(ctx context.Context, tx *sql.Tx, userId int32) (*model.WebAuthnUser, error) {
	var u model.User
	row := tx.QueryRowContext(ctx, `
	SELECT id, username
	FROM users
	WHERE id = ?;`, userId)
	err := row.Scan(&u.Id, &u.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
	SELECT credential_data
	FROM webauthn_credentials
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot query credential data: %w", err)
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var credData []byte
		if err := rows.Scan(&credData); err != nil {
			return nil, fmt.Errorf("cannot scan credential data: %w", err)
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(credData, &cred); err != nil {
			return nil, fmt.Errorf("cannot unmarshal credential data: %w", err)
		}
		credentials = append(credentials, cred)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate over credential rows: %w", err)
	}

	return &model.WebAuthnUser{
		Id:          u.Id,
		Name:        u.Username,
		DisplayName: u.Username,
		Icon:        "",
		Credentials: credentials,
	}, nil
}

func (store *AuthStore) UpdateWebAuthnCredentialAfterLogin(ctx context.Context, tx *sql.Tx, userId int32, credential *webauthn.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("cannot marshal credential data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE webauthn_credentials
	SET credential_data = ?, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = ?;
	`, data, userId)
	if err != nil {
		return fmt.Errorf("cannot update credential data: %w", err)
	}

	return nil
}

func (store *AuthStore) DeleteWebAuthnCredentials(ctx context.Context, tx *sql.Tx, userId int32) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM webauthn_credentials
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return fmt.Errorf("cannot delete credential data: %w", err)
	}

	return nil
}

func (store *AuthStore) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

func initdb(ctx context.Context, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := initUserTable(ctx, tx); err != nil {
		return err
	}
	if err := initPasswordTable(ctx, tx); err != nil {
		return err
	}
	if err := initWebAuthnTable(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}
	return nil
}

func initUserTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return fmt.Errorf("cannot create users table: %w", err)
	}
	return nil
}

func initPasswordTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS passwords (
		user_id INTEGER PRIMARY KEY,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`)
	if err != nil {
		return fmt.Errorf("cannot create passwords table: %w", err)
	}
	return nil
}

func initWebAuthnTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS webauthn_sessions (
		challenge TEXT PRIMARY KEY,
		session_data BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("cannot create webauthn table: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS webauthn_credentials (
		user_id INTEGER NOT NULL,
		credential_data BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`)
	if err != nil {
		return fmt.Errorf("cannot create webauthn credentials table: %w", err)
	}

	return nil
}

func randomIdBytes() ([]byte, error) {
	id := make([]byte, 8)
	_, err := rand.Read(id)
	if err != nil {
		return nil, fmt.Errorf("cannot generate random id: %w", err)
	}
	return id, nil
}

func userHandlerIDBytes(userId int32) []byte {
	return []byte(strconv.FormatInt(int64(userId), 10))
}
