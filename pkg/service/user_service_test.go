package service

import (
	"errors"
	"testing"

	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/pkg/store"
	"github.com/elug3/gochat/pkg/store/user/sqlite"
)

func newTestUserService() (*UserService, error) {
	store, err := sqlite.NewUserStore(&config.Config{
		NoSave: true,
	})
	if err != nil {
		return nil, err
	}
	s, err := NewUserService(store)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func TestUserService_Register(t *testing.T) {
	testCases := map[string]struct {
		username string
		password string
		wantErr  error
	}{
		"nornal": {username: "test", password: "password"},
	}
	s, err := newTestUserService()
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			user, err := s.Register(t.Context(), tc.username, tc.password)
			if err != nil {
				if errors.Is(err, tc.wantErr) {
					return
				}
				t.Fatalf("unexpected error: want: %q, but got: %q", tc.wantErr, err)
			}
			if user.Username != tc.username {
				t.Errorf("unexpected username: want: %q, but got: %q", tc.username, user.Username)
			}
		})
	}
}

func TestUserService_Auth(t *testing.T) {
	s, err := newTestUserService()
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		usernameIn   string
		passwordIn   string
		wantUsername string
		wantErr      error
	}{
		"normal":             {usernameIn: "test", passwordIn: "password", wantUsername: "test"},
		"incorrect password": {usernameIn: "test", passwordIn: "no secrets", wantErr: store.ErrUnauthorized},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := s.Login(t.Context(), tc.usernameIn, tc.passwordIn)
			if err != nil {
				if errors.Is(err, tc.wantErr) {
					t.SkipNow()
				}
				t.Errorf("unexpected error: want: %v, but got %v", tc.wantErr, err)
			} else if tc.wantErr != nil {
				t.Errorf("unexpected result: %v", result)
			}
		})

	}
}
