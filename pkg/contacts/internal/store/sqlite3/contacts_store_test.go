package sqlite3_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/elug3/gochat/pkg/contacts/internal/store"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
)

func NewTestContactsStore(t *testing.T) *sqlite3.ContactsStore {
	t.Helper()
	store, err := sqlite3.NewContactsStore("", true)
	if err != nil {
		t.Fatalf("failed to create ContactsStore: %v", err)
	}
	return store
}

func TestProfileCRUD(t *testing.T) {
	store := NewTestContactsStore(t)

	userId := int32(1)

	txc, err := store.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer txc.Rollback()

	// Test CreateProfile
	createdProfile, err := txc.CreateProfile(userId, "test profile", "")
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
	fetchedProfile, err := txc.GetProfile(userId)
	if err != nil {
		t.Errorf("failed to get profile: %v", err)
	}

	if fetchedProfile.Id != createdProfile.Id || fetchedProfile.Name != createdProfile.Name {
		t.Errorf("fetched profile does not match created profile")
	}

	// Test Update Profile
	updatedName := "updated profile"
	updatedProfile, err := txc.UpdateProfile(userId, updatedName, "")
	if err != nil {
		t.Errorf("failed to update profile: %v", err)
	}

	if updatedProfile.Id != userId || updatedProfile.Name != updatedName {
		t.Errorf("updated profile does not match expected values")
	}

	// Test DeleteProfile
	err = txc.DeleteProfile(userId)
	if err != nil {
		t.Errorf("failed to delete profile: %v", err)
	}

	_, err = txc.GetProfile(userId)
	if err == nil {
		t.Errorf("expected error when fetching deleted profile, got none")
	}
}

func TestContactCRUD(t *testing.T) {
	store := NewTestContactsStore(t)

	txc, err := store.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer txc.Rollback()

	user := generateTestProfile(t, txc)
	friend := generateTestProfile(t, txc)

	// create contact
	createdContact, err := txc.CreateContact(user.Id, friend.Id, nil)
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
	contacts, err := txc.ListUserContacts(user.Id)
	if err != nil {
		t.Errorf("failed to list contacts: %v", err)
	}
	if len(contacts) != 1 {
		t.Errorf("expected 1 contact, got %d", len(contacts))
	}

	// delete contact
	if err = txc.DeleteUserContact(user.Id, friend.Id); err != nil {
		t.Errorf("failed to delete contact: %s", err)
	}
}

var testUserIdCount int32 = 1000

func generateTestUserId() int32 {
	return atomic.AddInt32(&testUserIdCount, 1)
}

func generateTestProfile(t *testing.T, txc store.TxContacts) *model.Profile {
	t.Helper()

	p, err := txc.CreateProfile(generateTestUserId(), "Test User", "")
	if err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}

	return p
}

func generateTestGroup() {
}

func checkProfileExists(t *testing.T, txc store.TxContacts, userId int32) error {
	t.Helper()
	_, err := txc.GetProfile(userId)
	if err != nil {
		return err
	}
	return nil
}

func TestDeleteProfile(t *testing.T) {

	t.Run("delete existing profile", func(t *testing.T) {
		store := NewTestContactsStore(t)

		txc, err := store.Begin()
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}
		defer txc.Rollback()

		createdProfile := generateTestProfile(t, txc)

		err = txc.DeleteProfile(createdProfile.Id)
		if err != nil {
			t.Fatalf("failed to delete profile: %v", err)
		}

		if err := checkProfileExists(t, txc, createdProfile.Id); !errors.Is(err, errs.ErrNotFound) {
			t.Errorf("expected profile to be deleted, but it still exists")
		}
	})

	t.Run("delete non-existing profile", func(t *testing.T) {
		store := NewTestContactsStore(t)

		txc, err := store.Begin()
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}
		defer txc.Rollback()

		nonExistentUserId := int32(0)
		err = txc.DeleteProfile(nonExistentUserId)
		if err != nil {
			t.Fatalf("failed to delete non-existing profile: %v", err)
		}
	})
}
