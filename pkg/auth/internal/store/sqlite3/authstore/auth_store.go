package authstore

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/go-webauthn/webauthn/webauthn"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type ApiTokenOptions struct {
	Name     string
	Prefix   string
	Scopes   []string
	ExpireIn int64 // seconds
}

func NewAuthStore(ctx context.Context, db *sql.DB) (*Store, error) {
	if err := initdb(ctx, db); err != nil {
		return nil, fmt.Errorf("cannot initialize database: %w", err)
	}

	return &Store{db: db}, nil
}

func (store *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return store.db.BeginTx(ctx, opts)
}

func (store *Store) DB() *sql.DB {
	if store.db == nil {
		panic("database not initialized")
	}
	return store.db
}

func (store *Store) CreateUser(ctx context.Context, tx *sql.Tx, username string) (*model.User, error) {
	var u model.User
	err := tx.QueryRowContext(ctx, `
	INSERT INTO users (username)
	VALUES (?)
	RETURNING id, username;
	`, username).Scan(&u.Id, &u.Username)
	if err != nil {
		return nil, fmt.Errorf("cannot insert user: %w", handleSqlError(err))
	}

	return &u, nil
}

func (store *Store) GetUserById(ctx context.Context, tx *sql.Tx, userId int32) (*model.User, error) {
	var u model.User
	err := tx.QueryRowContext(ctx, `
	SELECT id, username
	FROM users
	WHERE id = ?;`, userId).Scan(&u.Id, &u.Username)
	if err != nil {
		return nil, fmt.Errorf("cannot get user: %w", handleSqlError(err))
	}
	return &u, nil
}

func (store *Store) SetPasswordHash(ctx context.Context, tx *sql.Tx, userId int32, passwordHash string) error {
	_, err := tx.ExecContext(ctx, `
	INSERT INTO passwords (user_id, password_hash)
	VALUES (?, ?)
	ON CONFLICT(user_id) DO UPDATE SET password_hash=excluded.password_hash, updated_at=CURRENT_TIMESTAMP;
	`, userId, passwordHash)
	if err != nil {
		return fmt.Errorf("cannot update password hash: %w", handleSqlError(err))
	}

	return nil
}

func (store *Store) GetPasswordByUsername(ctx context.Context, tx *sql.Tx, username string) (*model.PasswordCredential, error) {
	var pw model.PasswordCredential
	err := tx.QueryRowContext(ctx, `
	SELECT p.user_id, u.username, p.password_hash
	FROM passwords as p
	JOIN users as u ON u.id = p.user_id
	WHERE u.username = ?;`, username).Scan(&pw.UserId, &pw.Username, &pw.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("cannot get password: %w", handleSqlError(err))
	}
	return &pw, nil
}

func (store *Store) SaveSession(ctx context.Context, tx *sql.Tx, userId int32, sessionHash string, ip net.IP, userAgent string, createdAt, expiresAt time.Time) error {

	_, err := tx.ExecContext(ctx, `
	INSERT INTO sessions (session_hash, user_id, ip, user_agent, created_at, expires_at)
	VALUES (?, ?, ?, ?, ?, ?);
	`, sessionHash, userId, ip.String(), userAgent, createdAt, expiresAt)
	if err != nil {
		return fmt.Errorf("cannot insert session: %w", handleSqlError(err))
	}
	return nil
}

func (store *Store) GetSessionByHash(ctx context.Context, tx *sql.Tx, sessionHash string) (*model.Session, error) {
	var (
		session model.Session
		ip      string
		revoked sql.NullTime
	)
	err := tx.QueryRowContext(ctx, `
	SELECT session_hash, user_id, ip, user_agent, created_at, expires_at, revoked_at
	FROM sessions
	WHERE session_hash = ?;`, sessionHash).Scan(
		&session.SessionId,
		&session.UserId,
		&ip,
		&session.UserAgent,
		&session.CreatedAt,
		&session.ExpiresAt,
		&revoked,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get session: %w", handleSqlError(err))
	}
	session.IP = net.ParseIP(ip)
	if revoked.Valid {
		session.RevokedAt = revoked.Time
	}
	return &session, nil
}

func (store *Store) SaveWebauthnSessionData(ctx context.Context, tx *sql.Tx, session *webauthn.SessionData) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("cannot marshal session data: %w", handleSqlError(err))
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO webauthn_sessions (challenge, session_data)
	VALUES (?, ?);
	`, session.Challenge, data)
	if err != nil {
		return fmt.Errorf("cannot insert session data: %w", handleSqlError(err))
	}

	return nil
}

func (store *Store) GetWebauthnSessionData(ctx context.Context, tx *sql.Tx, challenge string) (*webauthn.SessionData, error) {
	row := tx.QueryRowContext(ctx, `
	SELECT session_data
	FROM webauthn_sessions
	WHERE challenge = ?;
	`, challenge)

	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, fmt.Errorf("cannot scan session data: %w", handleSqlError(err))
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("cannot unmarshal session data: %w", err)
	}

	return &session, nil
}

func (store *Store) DeleteWebauthnSessionData(ctx context.Context, tx *sql.Tx, challenge string) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM webauthn_sessions
	WHERE challenge = ?;
	`, challenge)
	if err != nil {
		return fmt.Errorf("cannot delete session data: %w", handleSqlError(err))
	}

	return nil
}

func (store *Store) SaveWebAuthnCredential(ctx context.Context, tx *sql.Tx, userId int32, credential_name string, credential *webauthn.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("cannot marshal credential data: %w", handleSqlError(err))
	}

	_, err = tx.ExecContext(ctx, `
	INSERT INTO webauthn_credentials (user_id, name, credential_data)
	VALUES (?, ?, ?);
	`, userId, credential_name, data)
	if err != nil {
		return fmt.Errorf("cannot insert credential data: %w", handleSqlError(err))
	}

	return nil
}

// GetWebAuthnUser returns the WebAuthnUser for the given userId.
// It returns empty credentials user for valid user with no credentials.
// If user does not exist, returns errs.ErrNotFound.
func (store *Store) GetWebAuthnUser(ctx context.Context, tx *sql.Tx, userId int32) (*model.WebAuthnUser, error) {
	var u model.User
	row := tx.QueryRowContext(ctx, `
	SELECT id, username
	FROM users
	WHERE id = ?;`, userId)
	err := row.Scan(&u.Id, &u.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", handleSqlError(err))
	}
	rows, err := tx.QueryContext(ctx, `
	SELECT credential_data
	FROM webauthn_credentials
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot query credential data: %w", handleSqlError(err))
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var credData []byte
		if err := rows.Scan(&credData); err != nil {
			return nil, fmt.Errorf("cannot scan credential data: %w", handleSqlError(err))
		}
		var cred webauthn.Credential
		if err := json.Unmarshal(credData, &cred); err != nil {
			return nil, fmt.Errorf("cannot unmarshal credential data: %w", err)
		}
		credentials = append(credentials, cred)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate over credential rows: %w", handleSqlError(err))
	}

	return &model.WebAuthnUser{
		Id:          u.Id,
		Name:        u.Username,
		DisplayName: u.Username,
		Icon:        "",
		Credentials: credentials,
	}, nil
}

func (store *Store) UpdateWebAuthnCredentialAfterLogin(ctx context.Context, tx *sql.Tx, userId int32, credential *webauthn.Credential) error {
	data, err := json.Marshal(credential)
	if err != nil {
		return fmt.Errorf("cannot marshal credential data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE webauthn_credentials
	SET credential_data = ?, last_used_at = CURRENT_TIMESTAMP
	WHERE user_id = ?;
	`, data, userId)
	if err != nil {
		return fmt.Errorf("cannot update credential data: %w", handleSqlError(err))
	}

	return nil
}

func (store *Store) DeleteWebAuthnCredentials(ctx context.Context, tx *sql.Tx, userId int32) error {
	_, err := tx.ExecContext(ctx, `
	DELETE FROM webauthn_credentials
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return fmt.Errorf("cannot delete credential data: %w", handleSqlError(err))
	}

	return nil
}

func (store *Store) UpdatePasskey(ctx context.Context, tx *sql.Tx, passkeyId int32, passkeyName string) (*model.Passkey, error) {
	row := tx.QueryRowContext(ctx, `
	UPDATE webauthn_credentials
	SET name = ?
	WHERE id = ?
	RETURNING id, name, user_id, created_at, last_used_at;
	`, passkeyName, passkeyId)

	var updatedPasskey model.Passkey
	if err := row.Scan(
		&updatedPasskey.Id,
		&updatedPasskey.KeyName,
		&updatedPasskey.UserId,
		&updatedPasskey.CreatedAt,
		&updatedPasskey.LastUsedAt,
	); err != nil {
		return nil, fmt.Errorf("cannot update passkey: %w", handleSqlError(err))
	}
	return &updatedPasskey, nil
}

func (store *Store) DeletePasskeyById(ctx context.Context, tx *sql.Tx, passkeyId int32) (*model.Passkey, error) {
	row := tx.QueryRowContext(ctx, `
	DELETE FROM webauthn_credentials
	WHERE id = ?
	RETURNING id, name, user_id, created_at, last_used_at;
	`, passkeyId)

	var deletedPasskey model.Passkey
	if err := row.Scan(
		&deletedPasskey.Id,
		&deletedPasskey.KeyName,
		&deletedPasskey.UserId,
		&deletedPasskey.CreatedAt,
		&deletedPasskey.LastUsedAt,
	); err != nil {
		return nil, fmt.Errorf("cannot delete passkey: %w", handleSqlError(err))
	}
	return &deletedPasskey, nil
}

func (store *Store) GetPasskeysByUserId(ctx context.Context, tx *sql.Tx, userId int32) ([]model.Passkey, error) {
	rows, err := tx.QueryContext(ctx, `
	SELECT id, name, user_id, created_at, last_used_at
	FROM webauthn_credentials
	WHERE user_id = ?;
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot query passkeys: %w", handleSqlError(err))
	}
	defer rows.Close()

	passkeys := make([]model.Passkey, 0)
	for rows.Next() {
		var pk model.Passkey
		if err := rows.Scan(
			&pk.Id,
			&pk.KeyName,
			&pk.UserId,
			&pk.CreatedAt,
			&pk.LastUsedAt); err != nil {
			return nil, fmt.Errorf("cannot scan passkey: %w", handleSqlError(err))
		}
		passkeys = append(passkeys, pk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate over passkey rows: %w", handleSqlError(err))
	}

	return passkeys, nil
}

func (store *Store) UpdatePasskeyLastUsedAt(ctx context.Context, tx *sql.Tx, passkeyId int32) error {
	_, err := tx.ExecContext(ctx, `
	UPDATE webauthn_credentials
	SET last_used_at = CURRENT_TIMESTAMP
	WHERE id = ?;
	`, passkeyId)
	if err != nil {
		return fmt.Errorf("cannot update passkey last_used_at: %w", handleSqlError(err))
	}
	return nil
}

func (store *Store) Close() error {
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
	if err := initSessionTable(ctx, tx); err != nil {
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

func initSessionTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS sessions (
		session_hash TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		ip TEXT NOT NULL,
		user_agent TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		revoked_at DATETIME,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`)
	if err != nil {
		return fmt.Errorf("cannot create sessions table: %w", err)
	}
	return nil
}

func initApiTokenTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS api_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token_hint TEXT NOT NULL UNIQUE,
		token_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		scopes JSON NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		revoked_at DATETIME,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE;
	);`)
	if err != nil {
		return fmt.Errorf("cannot create api_tokens table: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_api_tokens_token_prefix ON api_tokens(token_prefix);
	`); err != nil {
		return fmt.Errorf("cannot create api_tokens indexes: %w", err)
	}

	return nil
}

func initWebAuthnTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS webauthn_sessions (
		challenge TEXT PRIMARY KEY,
		session_data BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return fmt.Errorf("cannot create webauthn_session table: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS webauthn_credentials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		credential_data BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`)
	if err != nil {
		return fmt.Errorf("cannot create webauthn_credentials table: %w", err)
	}

	return nil
}

func handleSqlError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}

	var sqlErr sqlite3.Error
	if errors.As(err, &sqlErr) {
		if sqlErr.Code == sqlite3.ErrConstraint {
			switch sqlErr.ExtendedCode {
			case sqlite3.ErrConstraintUnique, sqlite3.ErrConstraintPrimaryKey:
				return errs.ErrExists
			default:
				return errs.ErrConstraint
			}
		}
	}

	return err
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

func hashApiToken(token string) (string, error) {
	hasher := sha512.New512_256()
	_, err := hasher.Write([]byte(token))
	if err != nil {
		return "", fmt.Errorf("cannot hash api token: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil

}
