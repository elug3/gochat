package sqlite3_test

import (
	"database/sql"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
)

func NewTestContactsStore(t *testing.T) *sqlite3.ContactsStore {
	t.Helper()
	store, _, err := sqlite3.NewContactsStore("", true)
	if err != nil {
		t.Fatalf("failed to create ContactsStore: %v", err)
	}
	return store
}

func TestProfileCRUD(t *testing.T) {
	store := NewTestContactsStore(t)

	userId := int32(1)

	tx, err := store.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Test CreateProfile
	createdProfile, err := store.CreateProfile(tx, userId, "test profile")
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}

	if createdProfile.Id != userId {
		t.Errorf("expected profile ID %d, got %d", userId, createdProfile.Id)
	}
	if createdProfile.Name != "test profile" {
		t.Errorf("expected profile name 'test profile', got '%s'", createdProfile.Name)
	}

	// Test GetProfile
	fetchedProfile, err := store.GetProfile(tx, userId)
	if err != nil {
		t.Errorf("failed to get profile: %v", err)
	}

	if fetchedProfile.Id != createdProfile.Id || fetchedProfile.Name != createdProfile.Name {
		t.Errorf("fetched profile does not match created profile")
	}

	// Test Update Profile
	updatedName := "updated profile"
	updatedProfile, err := store.UpdateProfile(tx, userId, updatedName, "")
	if err != nil {
		t.Errorf("failed to update profile: %v", err)
	}

	if updatedProfile.Id != userId || updatedProfile.Name != updatedName {
		t.Errorf("updated profile does not match expected values")
	}

	// Test DeleteProfile
	err = store.DeleteProfile(tx, userId)
	if err != nil {
		t.Errorf("failed to delete profile: %v", err)
	}

	_, err = store.GetProfile(tx, userId)
	if err == nil {
		t.Errorf("expected error when fetching deleted profile, got none")
	}
}

func TestContactCRUD(t *testing.T) {
	store := NewTestContactsStore(t)

	tx, err := store.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	user := generateTestProfile(t, store, tx)
	friend := generateTestProfile(t, store, tx)

	// create contact
	createdContact, err := store.CreateContact(tx, user.Id, friend.Id, nil)
	if err != nil {
		t.Fatalf("failed to create contact: %v", err)
	}
	if createdContact.ProfileId != friend.Id {
		t.Errorf("unexpected contact profile ID: want: %d, but got: %d", friend.Id, createdContact.ProfileId)
	}
	if createdContact.Name != friend.Name {
		t.Errorf("unexpected contact name: want: %s, but got %s", friend.Name, createdContact.Name)
	}

	// TODO: add update contact

	// list contacts
	contacts, err := store.ListUserContacts(tx, user.Id)
	if err != nil {
		t.Errorf("failed to list contacts: %v", err)
	}
	if len(contacts) != 1 {
		t.Errorf("expected 1 contact, got %d", len(contacts))
	}

	// delete contact
	if err = store.DeleteUserContact(tx, user.Id, friend.Id); err != nil {
		t.Errorf("failed to delete contact: %s", err)
	}
}

var testUserIdCount int32 = 1000

func generateTestUserId() int32 {
	return atomic.AddInt32(&testUserIdCount, 1)
}

func generateTestProfile(t *testing.T, store *sqlite3.ContactsStore, tx *sql.Tx) *model.Profile {
	t.Helper()

	p, err := store.CreateProfile(tx, generateTestUserId(), "Test User")
	if err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}

	return p
}

func generateTestGroup() {
}

func checkProfileExists(t *testing.T, store *sqlite3.ContactsStore, tx *sql.Tx, userId int32) error {
	t.Helper()
	_, err := store.GetProfile(tx, userId)
	if err != nil {
		return err
	}
	return nil
}

func TestDeleteProfile(t *testing.T) {

	t.Run("delete existing profile", func(t *testing.T) {
		store := NewTestContactsStore(t)

		tx, err := store.Begin()
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		createdProfile := generateTestProfile(t, store, tx)

		err = store.DeleteProfile(tx, createdProfile.Id)
		if err != nil {
			t.Fatalf("failed to delete profile: %v", err)
		}

		if err := checkProfileExists(t, store, tx, createdProfile.Id); !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("expected profile to be deleted, but it still exists")
		}
	})

	t.Run("delete non-existing profile", func(t *testing.T) {
		store := NewTestContactsStore(t)

		tx, err := store.Begin()
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		nonExistentUserId := int32(0)
		err = store.DeleteProfile(tx, nonExistentUserId)
		if !errors.Is(err, errs.ErrUserNotExists) {
			t.Fatalf("expected ErrUserNotExists for missing profile, got %v", err)
		}
	})
}
