package contacts_test

import (
	"errors"
	"testing"

	"github.com/elug3/gochat/pkg/contacts"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
)

var testOptions = &contacts.Options{
	NoSave:  true,
	NoEvent: true,
	NoIcons: true,
}

func NewTestService(t *testing.T) *contacts.ContactsService {
	t.Helper()
	s, err := contacts.NewContactsService(testOptions)
	if err != nil {
		t.Fatalf("failed to create contacts service: %v", err)
	}
	return s
}

func TestAccess(t *testing.T) {
}

var (
	SeqUserId int32 = 1000
)

func genUserId() int32 {
	SeqUserId = SeqUserId + 1
	return SeqUserId
}

func newTestProfile(t *testing.T, s *contacts.ContactsService, name string) *model.Profile {
	t.Helper()
	pid := genUserId()

	p, err := s.CreateProfile(t.Context(), pid, name)
	if err != nil {
		t.Fatalf("cannot create test profile: %v", err)
	}

	if p.Id != pid {
		t.Fatalf("expected profile id %d, but got %d", pid, p.Id)
	}
	if p.Name != name {
		t.Fatalf("expected profile name %q, but got %q", name, p.Name)
	}
	return p
}

func newTestGroup(t *testing.T, s *contacts.ContactsService, pid int32, name string) *model.Group {
	t.Helper()

	g, err := s.CreateUserGroup(pid, name)
	if err != nil {
		t.Fatalf("cannot create test group: %v", err)
	}

	if g.Name != name {
		t.Fatalf("expected group name %q, but got %q", name, g.Name)
	}

	return g
}

func deleteTestProfile(t *testing.T, s *contacts.ContactsService, pid int32) {
	t.Helper()

	if err := s.DeleteProfile(pid); err != nil {
		t.Fatalf("failed to delete profile %q: %v", pid, err)
	}
}

func newTestContact(t *testing.T, s *contacts.ContactsService, ownerId int32, targetId int32) *model.Contact {
	t.Helper()

	c, err := s.AddToContacts(ownerId, targetId)
	if err != nil {
		t.Fatalf("cannot create test contact: %v", err)
	}

	if c.ProfileId != targetId {
		t.Fatalf("expected contact target id %d, but got %d", targetId, c.ProfileId)
	}

	return c
}

func TestProfile(t *testing.T) {
	t.Run("Profile CRUD", func(t *testing.T) {
		service := NewTestService(t)

		// create
		createdProfile := newTestProfile(t, service, "test user")

		// delete
		deleteTestProfile(t, service, createdProfile.Id)
	})

	t.Run("delete profile with group", func(t *testing.T) {
		t.Run("cannot delete profile with group owner", func(t *testing.T) {
			service := NewTestService(t)

			p := newTestProfile(t, service, "owner")
			newTestGroup(t, service, p.Id, "test group")

			err := service.DeleteProfile(p.Id)
			if err == nil {
				t.Fatalf("expected error when deleting profile with group owner, but got none")
			}
			if !errors.Is(err, errs.ErrPermissionDenied) {
				t.Fatalf("expected %v, but got %v", errs.ErrPermissionDenied, err)
			}
		})
	})
	t.Run("delete profile with contact", func(t *testing.T) {
		t.Run("deletes contact when profile is deleted", func(t *testing.T) {
			service := NewTestService(t)
			owner := newTestProfile(t, service, "test profile")
			friend := newTestProfile(t, service, "friend profile")

			newTestContact(t, service, owner.Id, friend.Id)

			err := service.DeleteProfile(friend.Id)
			if err != nil {
				t.Fatalf("failed to delete profile with contact: %v", err)
			}

			// check contact is also deleted
			contacts, err := service.ListUserContacts(owner.Id)
			if err != nil {
				t.Fatalf("failed to list contacts: %v", err)
			}
			if len(contacts) != 0 {
				t.Fatalf("expected 0 contacts after deleting profile, but got %d", len(contacts))
			}
		})
	})
}

func TestGroup(t *testing.T) {
}
