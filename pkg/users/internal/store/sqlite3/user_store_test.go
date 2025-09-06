package sqlite3

import (
	"errors"
	"testing"

	"github.com/elug3/gochat/pkg/users/internal/errs"
)

func TestUserStore_CreateGetUpdateDelete(t *testing.T) {
	// use in-memory database for tests
	store, err := NewUserStore()
	if err != nil {
		t.Fatalf("failed to create new user store: %v", err)
	}

	tx, err := store.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	// rollback transaction to avoid side-effects
	defer tx.Rollback()

	// Test CreateUser
	username := "testuser"
	user, err := tx.CreateUser(username)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if user.Username != username {
		t.Fatalf("expected username %q, got %q", username, user.Username)
	}

	// Test GetUser
	dbUser, err := tx.GetUser(user.Id)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if dbUser.Username != username {
		t.Fatalf("expected username %q, got %q", username, dbUser.Username)
	}

	// Test UpdateUser
	newUsername := "updatedUser"
	updatedUser, err := tx.UpdateUser(user.Id, newUsername)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}
	if updatedUser.Username != newUsername {
		t.Fatalf("expected updated username %q, got %q", newUsername, updatedUser.Username)
	}

	// Test DeleteUser
	if err := tx.DeleteUser(user.Id); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
	_, err = tx.GetUser(user.Id)
	if err == nil {
		t.Fatal("expected error when retrieving deleted user, got nil")
	}
	if !errors.Is(err, errs.ErrNotFound) {
		t.Fatalf("expected error %v, got %v", errs.ErrNotFound, err)
	}
}
