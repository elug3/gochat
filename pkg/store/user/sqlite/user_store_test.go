package sqlite

import (
	"errors"
	"testing"

	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/pkg/store"
	_ "github.com/mattn/go-sqlite3"
)

func NewTestStore() (*UserStore, error) {
	store, err := NewUserStore(&config.Config{
		NoSave: true,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func TestUserStore_CreateUser(t *testing.T) {
	userStore, err := NewTestStore()
	if err != nil {
		t.Fatal(err)
	}
	txc, err := userStore.Begin()
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		username     string
		wantUsername string
		wantErr      error
	}{
		"normal":         {username: "test", wantUsername: "test"},
		"empty username": {username: "", wantUsername: "", wantErr: store.ErrBadRequest},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := txc.CreateUser(tc.username)
			if err != nil {
				if errors.Is(err, tc.wantErr) {
					t.SkipNow()
				}
				t.Fatalf("could not create user: %q: %v", tc.username, err)
			}
			if result.Username != tc.wantUsername {
				t.Errorf("expected username %q, but got %q", tc.wantUsername, result.Username)
			}
		})
	}
}
