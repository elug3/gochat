package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/internal/model"
)

func testStore() (*AuthStore, error) {
	store, err := NewAuthStore("", true)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// testUser is a helper function to create a test user in the AuthStore.
func testUser(ctx context.Context, store *AuthStore, tx *sql.Tx, username string) (*model.User, error) {
	return store.CreateUser(ctx, tx, username)
}

func testPassword(ctx context.Context, store *AuthStore, tx *sql.Tx, uid int32, password string) error {
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

func TestNewUser(t *testing.T) {
	store, err := testStore()
	if err != nil {
		t.Fatalf("cannot create test store: %v", err)
	}
	defer store.Close()

	testCases := map[string]struct {
		username string
		wantErr  error
	}{
		"create user": {username: "testuser", wantErr: nil},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tx, err := store.db.Begin()
			if err != nil {
				t.Fatalf("cannot begin transaction: %v", err)
			}
			defer tx.Rollback()

			u, err := store.CreateUser(t.Context(), tx, tc.username)
			if errors.Is(err, tc.wantErr) == false {
				t.Errorf("NewUser(%q) error = %v, wantErr %v", tc.username, err, tc.wantErr)
			}
			if err == nil && u.Username != tc.username {
				t.Errorf("NewUser(%q) = %v, want username %v", tc.username, u.Username, tc.username)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	store, err := testStore()
	if err != nil {
		t.Fatalf("cannot create test store")
	}

	testCases := map[string]struct {
		password string
		wantErr  error
	}{}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tx, err := store.db.Begin()
			if err != nil {
				t.Fatalf("cannot begin transaction: %v", err)
			}
			defer tx.Rollback()

			u, err := testUser(t.Context(), store, tx, "testuser")
			if err != nil {
				t.Fatalf("cannot create user: %v", err)
			}
			err = testPassword(t.Context(), store, tx, u.Id, tc.password)
			if err != tc.wantErr {
				t.Errorf("expect error %v. but got: %v", tc.wantErr, err)
			}

		})

	}
}
