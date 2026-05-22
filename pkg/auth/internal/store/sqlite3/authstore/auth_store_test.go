package authstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/domain"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/go-webauthn/webauthn/webauthn"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewAuthStore(ctx, db)
	if err != nil {
		t.Fatalf("cannot create test store: %v", err)
	}
	return store
}

func beginTx(t *testing.T, store *Store) *sql.Tx {
	t.Helper()

	tx, err := store.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("cannot begin transaction: %v", err)
	}
	return tx
}

func mustCreateUser(t *testing.T, store *Store, tx *sql.Tx, username string) *domain.User {
	t.Helper()

	u, err := store.CreateUser(t.Context(), tx, username)
	if err != nil {
		t.Fatalf("cannot create user: %v", err)
	}
	return u
}

func testPassword(ctx context.Context, store *Store, tx *sql.Tx, uid int32, password string) error {
	passwordHash := newHash(password)
	return store.SetPasswordHash(ctx, tx, uid, passwordHash)
}

func newHash(password string) string {
	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		panic(err)
	}
	return passwordHash
}

func testCredential(id string) *webauthn.Credential {
	return &webauthn.Credential{
		ID:              []byte(id),
		PublicKey:       []byte("pk-" + id),
		AttestationType: "none",
	}
}

func testSessionData(challenge string) *webauthn.SessionData {
	return &webauthn.SessionData{
		Challenge:            challenge,
		RelyingPartyID:       "example.com",
		UserID:               []byte("user-id"),
		AllowedCredentialIDs: [][]byte{[]byte("cred-1")},
		Expires:              time.Now().Add(5 * time.Minute),
	}
}

func timesClose(a, b time.Time) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Second
}

func TestNewAuthStore(t *testing.T) {
	store, err := NewAuthStore("", true)
	if err != nil {
		t.Fatalf("cannot create test store: %v", err)
	}
	if store.DB() == nil {
		t.Fatalf("expected database handle")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("cannot close store: %v", err)
	}
}

func TestBeginTx(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx, err := store.BeginTx(t.Context(), &sql.TxOptions{})
	if err != nil {
		t.Fatalf("cannot begin transaction: %v", err)
	}
	if tx == nil {
		t.Fatalf("expected transaction")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("cannot rollback transaction: %v", err)
	}
}

func TestDB(t *testing.T) {
	t.Run("initialized", func(t *testing.T) {
		store := newTestStore(t)
		defer store.Close()

		if store.DB() == nil {
			t.Fatalf("expected database handle")
		}
	})

	t.Run("nil panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic for nil database")
			}
		}()
		store := &Store{}
		_ = store.DB()
	})
}

func TestCreateUser(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u, err := store.CreateUser(t.Context(), tx, "testuser")
	if err != nil {
		t.Fatalf("CreateUser error: %v", err)
	}
	if u.Username != "testuser" {
		t.Fatalf("CreateUser username = %q, want %q", u.Username, "testuser")
	}
	if u.Id == 0 {
		t.Fatalf("CreateUser id = 0, want non-zero")
	}

	if _, err := store.CreateUser(t.Context(), tx, "testuser"); err == nil {
		t.Fatalf("expected duplicate username error")
	}
}

func TestGetUserById(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	got, err := store.GetUserById(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetUserById error: %v", err)
	}
	if got.Username != u.Username {
		t.Fatalf("GetUserById username = %q, want %q", got.Username, u.Username)
	}

	_, err = store.GetUserById(t.Context(), tx, u.Id+100)
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestSetPasswordHash(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	passwordHash := newHash("password123")
	if err := store.SetPasswordHash(t.Context(), tx, u.Id, passwordHash); err != nil {
		t.Fatalf("SetPasswordHash error: %v", err)
	}

	var gotHash string
	if err := tx.QueryRowContext(t.Context(), `
		SELECT password_hash
		FROM passwords
		WHERE user_id = ?;`, u.Id).Scan(&gotHash); err != nil {
		t.Fatalf("query password hash: %v", err)
	}
	if gotHash != passwordHash {
		t.Fatalf("password hash mismatch")
	}

	updatedHash := newHash("password456")
	if err := store.SetPasswordHash(t.Context(), tx, u.Id, updatedHash); err != nil {
		t.Fatalf("SetPasswordHash update error: %v", err)
	}
	if err := tx.QueryRowContext(t.Context(), `
		SELECT password_hash
		FROM passwords
		WHERE user_id = ?;`, u.Id).Scan(&gotHash); err != nil {
		t.Fatalf("query password hash: %v", err)
	}
	if gotHash != updatedHash {
		t.Fatalf("updated password hash mismatch")
	}
}

func TestGetPasswordByUsername(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	password := "password123"
	if err := testPassword(t.Context(), store, tx, u.Id, password); err != nil {
		t.Fatalf("cannot set password: %v", err)
	}

	pw, err := store.GetPasswordByUsername(t.Context(), tx, u.Username)
	if err != nil {
		t.Fatalf("GetPasswordByUsername error: %v", err)
	}
	if pw.UserId != u.Id {
		t.Fatalf("password user id = %d, want %d", pw.UserId, u.Id)
	}
	if pw.Username != u.Username {
		t.Fatalf("password username = %q, want %q", pw.Username, u.Username)
	}
	if !pw.ValidatePassword(password) {
		t.Fatalf("expected password to match")
	}

	_, err = store.GetPasswordByUsername(t.Context(), tx, "missing-user")
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestSaveSession(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	ip := net.ParseIP("192.0.2.1")
	createdAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := createdAt.Add(2 * time.Hour)
	sessionHash := "session-hash"
	userAgent := "test-agent"

	if err := store.SaveSession(t.Context(), tx, u.Id, sessionHash, ip, userAgent, createdAt, expiresAt); err != nil {
		t.Fatalf("SaveSession error: %v", err)
	}

	var (
		gotHash    string
		gotUserId  int32
		gotIP      string
		gotAgent   string
		gotCreated time.Time
		gotExpires time.Time
	)
	if err := tx.QueryRowContext(t.Context(), `
		SELECT session_hash, user_id, ip, user_agent, created_at, expires_at
		FROM sessions
		WHERE session_hash = ?;`, sessionHash).Scan(
		&gotHash,
		&gotUserId,
		&gotIP,
		&gotAgent,
		&gotCreated,
		&gotExpires,
	); err != nil {
		t.Fatalf("query session: %v", err)
	}
	if gotHash != sessionHash {
		t.Fatalf("session hash = %q, want %q", gotHash, sessionHash)
	}
	if gotUserId != u.Id {
		t.Fatalf("session user id = %d, want %d", gotUserId, u.Id)
	}
	if gotIP != ip.String() {
		t.Fatalf("session ip = %q, want %q", gotIP, ip.String())
	}
	if gotAgent != userAgent {
		t.Fatalf("session user agent = %q, want %q", gotAgent, userAgent)
	}
	if !timesClose(gotCreated, createdAt) {
		t.Fatalf("session created_at mismatch")
	}
	if !timesClose(gotExpires, expiresAt) {
		t.Fatalf("session expires_at mismatch")
	}
}

func TestGetSessionByHash(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	ip := net.ParseIP("198.51.100.5")
	createdAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := createdAt.Add(3 * time.Hour)
	sessionHash := "session-hash"
	userAgent := "test-agent"

	if err := store.SaveSession(t.Context(), tx, u.Id, sessionHash, ip, userAgent, createdAt, expiresAt); err != nil {
		t.Fatalf("SaveSession error: %v", err)
	}

	session, err := store.GetSessionByHash(t.Context(), tx, sessionHash)
	if err != nil {
		t.Fatalf("GetSessionByHash error: %v", err)
	}
	if session.SessionId != sessionHash {
		t.Fatalf("session id = %q, want %q", session.SessionId, sessionHash)
	}
	if session.UserId != u.Id {
		t.Fatalf("session user id = %d, want %d", session.UserId, u.Id)
	}
	if !session.IP.Equal(ip) {
		t.Fatalf("session ip = %v, want %v", session.IP, ip)
	}
	if session.UserAgent != userAgent {
		t.Fatalf("session user agent = %q, want %q", session.UserAgent, userAgent)
	}
	if !timesClose(session.CreatedAt, createdAt) {
		t.Fatalf("session created_at mismatch")
	}
	if !timesClose(session.ExpiresAt, expiresAt) {
		t.Fatalf("session expires_at mismatch")
	}
	if !session.RevokedAt.IsZero() {
		t.Fatalf("expected zero revoked_at")
	}

	_, err = store.GetSessionByHash(t.Context(), tx, "missing-hash")
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestSaveWebauthnSessionData(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	session := testSessionData("challenge-1")
	if err := store.SaveWebauthnSessionData(t.Context(), tx, session); err != nil {
		t.Fatalf("SaveWebauthnSessionData error: %v", err)
	}

	var raw []byte
	if err := tx.QueryRowContext(t.Context(), `
		SELECT session_data
		FROM webauthn_sessions
		WHERE challenge = ?;`, session.Challenge).Scan(&raw); err != nil {
		t.Fatalf("query webauthn session: %v", err)
	}
	if len(raw) == 0 {
		t.Fatalf("expected session_data payload")
	}
	var got webauthn.SessionData
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal session_data: %v", err)
	}
	if got.Challenge != session.Challenge {
		t.Fatalf("challenge = %q, want %q", got.Challenge, session.Challenge)
	}
}

func TestGetWebauthnSessionData(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	session := testSessionData("challenge-2")
	if err := store.SaveWebauthnSessionData(t.Context(), tx, session); err != nil {
		t.Fatalf("SaveWebauthnSessionData error: %v", err)
	}

	got, err := store.GetWebauthnSessionData(t.Context(), tx, session.Challenge)
	if err != nil {
		t.Fatalf("GetWebauthnSessionData error: %v", err)
	}
	if got.Challenge != session.Challenge {
		t.Fatalf("challenge = %q, want %q", got.Challenge, session.Challenge)
	}
	if got.RelyingPartyID != session.RelyingPartyID {
		t.Fatalf("rpId = %q, want %q", got.RelyingPartyID, session.RelyingPartyID)
	}
	if !bytes.Equal(got.UserID, session.UserID) {
		t.Fatalf("user id mismatch")
	}
	if len(got.AllowedCredentialIDs) != len(session.AllowedCredentialIDs) {
		t.Fatalf("allowed credentials length mismatch")
	}
	if len(got.AllowedCredentialIDs) == 1 && !bytes.Equal(got.AllowedCredentialIDs[0], session.AllowedCredentialIDs[0]) {
		t.Fatalf("allowed credential id mismatch")
	}
	if !timesClose(got.Expires, session.Expires) {
		t.Fatalf("expires mismatch")
	}

	_, err = store.GetWebauthnSessionData(t.Context(), tx, "missing-challenge")
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteWebauthnSessionData(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	session := testSessionData("challenge-3")
	if err := store.SaveWebauthnSessionData(t.Context(), tx, session); err != nil {
		t.Fatalf("SaveWebauthnSessionData error: %v", err)
	}
	if err := store.DeleteWebauthnSessionData(t.Context(), tx, session.Challenge); err != nil {
		t.Fatalf("DeleteWebauthnSessionData error: %v", err)
	}
	_, err := store.GetWebauthnSessionData(t.Context(), tx, session.Challenge)
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestSaveWebAuthnCredential(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	cred := testCredential("cred-1")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", cred); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}

	var (
		name string
		data []byte
	)
	if err := tx.QueryRowContext(t.Context(), `
		SELECT name, credential_data
		FROM webauthn_credentials
		WHERE user_id = ?;`, u.Id).Scan(&name, &data); err != nil {
		t.Fatalf("query credential data: %v", err)
	}
	if name != "key-1" {
		t.Fatalf("credential name = %q, want %q", name, "key-1")
	}
	var got webauthn.Credential
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal credential: %v", err)
	}
	if !bytes.Equal(got.ID, cred.ID) {
		t.Fatalf("credential id mismatch")
	}
}

func TestGetWebAuthnUser(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	_, err := store.GetWebAuthnUser(t.Context(), tx, 999)
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}

	u := mustCreateUser(t, store, tx, "testuser")
	user, err := store.GetWebAuthnUser(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetWebAuthnUser error: %v", err)
	}
	if user.Name != u.Username {
		t.Fatalf("user name = %q, want %q", user.Name, u.Username)
	}
	if len(user.Credentials) != 0 {
		t.Fatalf("expected no credentials")
	}

	cred := testCredential("cred-1")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", cred); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	user, err = store.GetWebAuthnUser(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetWebAuthnUser error: %v", err)
	}
	if len(user.Credentials) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(user.Credentials))
	}
	if !bytes.Equal(user.Credentials[0].ID, cred.ID) {
		t.Fatalf("credential id mismatch")
	}
}

func TestUpdateWebAuthnCredentialAfterLogin(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	oldCred := testCredential("cred-old")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", oldCred); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	passkeys, err := store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if len(passkeys) != 1 {
		t.Fatalf("expected 1 passkey")
	}
	beforeLastUsed := passkeys[0].LastUsedAt

	newCred := testCredential("cred-new")
	if err := store.UpdateWebAuthnCredentialAfterLogin(t.Context(), tx, u.Id, newCred); err != nil {
		t.Fatalf("UpdateWebAuthnCredentialAfterLogin error: %v", err)
	}

	user, err := store.GetWebAuthnUser(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetWebAuthnUser error: %v", err)
	}
	if len(user.Credentials) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(user.Credentials))
	}
	if !bytes.Equal(user.Credentials[0].ID, newCred.ID) {
		t.Fatalf("updated credential id mismatch")
	}

	passkeys, err = store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if passkeys[0].LastUsedAt.Before(beforeLastUsed) {
		t.Fatalf("expected last_used_at to be updated")
	}
}

func TestDeleteWebAuthnCredentials(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", testCredential("cred-1")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-2", testCredential("cred-2")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	if err := store.DeleteWebAuthnCredentials(t.Context(), tx, u.Id); err != nil {
		t.Fatalf("DeleteWebAuthnCredentials error: %v", err)
	}
	user, err := store.GetWebAuthnUser(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetWebAuthnUser error: %v", err)
	}
	if len(user.Credentials) != 0 {
		t.Fatalf("expected 0 credentials")
	}
}

func TestUpdatePasskey(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", testCredential("cred-1")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	passkeys, err := store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if len(passkeys) != 1 {
		t.Fatalf("expected 1 passkey")
	}
	passkeyId := passkeys[0].Id

	updated, err := store.UpdatePasskey(t.Context(), tx, passkeyId, "key-updated")
	if err != nil {
		t.Fatalf("UpdatePasskey error: %v", err)
	}
	if updated.Id != passkeyId {
		t.Fatalf("updated passkey id = %d, want %d", updated.Id, passkeyId)
	}
	if updated.KeyName != "key-updated" {
		t.Fatalf("updated passkey name = %q, want %q", updated.KeyName, "key-updated")
	}

	passkeys, err = store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if passkeys[0].KeyName != "key-updated" {
		t.Fatalf("passkey name = %q, want %q", passkeys[0].KeyName, "key-updated")
	}

	if _, err := store.UpdatePasskey(t.Context(), tx, 9999, "missing"); err == nil {
		t.Fatalf("expected error for missing passkey")
	}
}

func TestDeletePasskeyById(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", testCredential("cred-1")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	passkeys, err := store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	passkeyId := passkeys[0].Id

	deleted, err := store.DeletePasskeyById(t.Context(), tx, passkeyId)
	if err != nil {
		t.Fatalf("DeletePasskeyById error: %v", err)
	}
	if deleted.Id != passkeyId {
		t.Fatalf("deleted passkey id = %d, want %d", deleted.Id, passkeyId)
	}

	passkeys, err = store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if len(passkeys) != 0 {
		t.Fatalf("expected 0 passkeys")
	}

	if _, err := store.DeletePasskeyById(t.Context(), tx, 9999); err == nil {
		t.Fatalf("expected error for missing passkey")
	}
}

func TestGetPasskeysByUserId(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	passkeys, err := store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if len(passkeys) != 0 {
		t.Fatalf("expected 0 passkeys")
	}

	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", testCredential("cred-1")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-2", testCredential("cred-2")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}

	passkeys, err = store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if len(passkeys) != 2 {
		t.Fatalf("expected 2 passkeys, got %d", len(passkeys))
	}
	seen := map[string]bool{}
	for _, pk := range passkeys {
		seen[pk.KeyName] = true
	}
	if !seen["key-1"] || !seen["key-2"] {
		t.Fatalf("expected both passkey names")
	}
}

func TestUpdatePasskeyLastUsedAt(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	tx := beginTx(t, store)
	defer tx.Rollback()

	u := mustCreateUser(t, store, tx, "testuser")
	if err := store.SaveWebAuthnCredential(t.Context(), tx, u.Id, "key-1", testCredential("cred-1")); err != nil {
		t.Fatalf("SaveWebAuthnCredential error: %v", err)
	}
	passkeys, err := store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	passkeyId := passkeys[0].Id
	before := passkeys[0].LastUsedAt

	if err := store.UpdatePasskeyLastUsedAt(t.Context(), tx, passkeyId); err != nil {
		t.Fatalf("UpdatePasskeyLastUsedAt error: %v", err)
	}
	passkeys, err = store.GetPasskeysByUserId(t.Context(), tx, u.Id)
	if err != nil {
		t.Fatalf("GetPasskeysByUserId error: %v", err)
	}
	if passkeys[0].LastUsedAt.Before(before) {
		t.Fatalf("expected last_used_at to be updated")
	}
}

func TestClose(t *testing.T) {
	store := newTestStore(t)
	if err := store.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}
