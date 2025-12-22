package contacts

import (
	"errors"
	"testing"

	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
)

func NewTestService(t *testing.T) *ContactsService {
	t.Helper()
	contactsStore, isNewStore, err := sqlite3.NewContactsStore("", true)
	if err != nil {
		t.Fatalf("failed to create contacts store: %v", err)
	}
	s, err := NewContactsService(ServiceDeps{
		Store:      contactsStore,
		IsNewStore: isNewStore,
	})
	if err != nil {
		t.Fatalf("failed to create contacts service: %v", err)
	}
	return s
}

func TestAccess(t *testing.T) {
	t.Run("invite uses acting user for permission checks", func(t *testing.T) {
		service := NewTestService(t)
		inviter := newTestProfile(t, service, "inviter")
		invitee := newTestProfile(t, service, "invitee")
		group := newTestGroup(t, service, inviter.Id, "group")

		can, _, err := service.Can(t.Context(), AccessRequest{
			UserId:   inviter.Id,
			ChatId:   group.Id,
			TargetId: invitee.Id,
			Action:   access.ActionInvite,
		})
		if err != nil {
			t.Fatalf("unexpected error while checking invite permission: %v", err)
		}
		if !can {
			t.Fatalf("expected inviter to be able to invite members into their group")
		}
	})

	t.Run("manager cannot delete owner, owner can delete manager", func(t *testing.T) {
		service := NewTestService(t)
		owner := newTestProfile(t, service, "owner")
		manager := newTestProfile(t, service, "manager")
		group := newTestGroup(t, service, owner.Id, "group")
		addMemberWithRole(t, service, group.Id, manager.Id, access.RoleManager)

		if err := service.DeleteMember(t.Context(), group.Id, manager.Id, owner.Id); !errors.Is(err, errs.ErrPermissionDenied) {
			t.Fatalf("expected manager deleting owner to be denied, got %v", err)
		}

		if err := service.DeleteMember(t.Context(), group.Id, owner.Id, manager.Id); err != nil {
			t.Fatalf("owner deleting manager should succeed, got %v", err)
		}
	})
}

var (
	SeqUserId int32 = 1000
)

func genUserId() int32 {
	SeqUserId = SeqUserId + 1
	return SeqUserId
}

func newTestProfile(t *testing.T, s *ContactsService, name string) *model.Profile {
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

func newTestGroup(t *testing.T, s *ContactsService, pid int32, name string) *model.Group {
	t.Helper()

	g, err := s.CreateUserGroup(t.Context(), pid, name)
	if err != nil {
		t.Fatalf("cannot create test group: %v", err)
	}

	if g.Name != name {
		t.Fatalf("expected group name %q, but got %q", name, g.Name)
	}

	return g
}

func addMemberWithRole(t *testing.T, s *ContactsService, groupId int, userId int32, role access.Role) {
	t.Helper()

	tx, err := s.store.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx for test member: %v", err)
	}
	defer tx.Rollback()

	if _, err := s.store.CreateMember(tx, groupId, userId, role); err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit test member: %v", err)
	}
}

func deleteTestProfile(t *testing.T, s *ContactsService, pid int32) {
	t.Helper()

	if err := s.DeleteProfile(t.Context(), pid); err != nil {
		t.Fatalf("failed to delete profile %q: %v", pid, err)
	}
}

func newTestContact(t *testing.T, s *ContactsService, ownerId int32, targetId int32) *model.Contact {
	t.Helper()

	c, err := s.AddToContacts(t.Context(), ownerId, targetId)
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

			err := service.DeleteProfile(t.Context(), p.Id)
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

			err := service.DeleteProfile(t.Context(), friend.Id)
			if err != nil {
				t.Fatalf("failed to delete profile with contact: %v", err)
			}

			// check contact is also deleted
			contacts, err := service.ListUserContacts(t.Context(), owner.Id)
			if err != nil {
				t.Fatalf("failed to list contacts: %v", err)
			}
			if len(contacts) != 0 {
				t.Fatalf("expected 0 contacts after deleting profile, but got %d", len(contacts))
			}
		})
	})
}

func TestContacts(t *testing.T) {
	t.Run("add and list contacts", func(t *testing.T) {
		service := NewTestService(t)

		owner := newTestProfile(t, service, "owner")
		friend := newTestProfile(t, service, "friend")

		contact, err := service.AddToContacts(t.Context(), owner.Id, friend.Id)
		if err != nil {
			t.Fatalf("failed to add contact: %v", err)
		}
		if contact.ProfileId != friend.Id {
			t.Fatalf("expected contact profile id %d, got %d", friend.Id, contact.ProfileId)
		}
		if contact.Name != friend.Name {
			t.Fatalf("expected contact name %q, got %q", friend.Name, contact.Name)
		}

		contacts, err := service.ListUserContacts(t.Context(), owner.Id)
		if err != nil {
			t.Fatalf("failed to list contacts: %v", err)
		}
		if len(contacts) != 1 {
			t.Fatalf("expected 1 contact, got %d", len(contacts))
		}
		if contacts[0].ProfileId != friend.Id {
			t.Fatalf("unexpected listed contact profile id, want %d got %d", friend.Id, contacts[0].ProfileId)
		}
	})

	t.Run("alias remains nil when not provided", func(t *testing.T) {
		service := NewTestService(t)

		owner := newTestProfile(t, service, "owner")
		friend := newTestProfile(t, service, "friend")

		contact := newTestContact(t, service, owner.Id, friend.Id)
		if contact.Alias != nil {
			t.Fatalf("expected alias to be nil when not provided, got %v", contact.Alias)
		}

		listed, err := service.ListUserContacts(t.Context(), owner.Id)
		if err != nil {
			t.Fatalf("failed to list contacts: %v", err)
		}
		if len(listed) != 1 {
			t.Fatalf("expected 1 contact, got %d", len(listed))
		}
		if listed[0].Alias != nil {
			t.Fatalf("expected alias to stay nil when reading contact, got %v", listed[0].Alias)
		}
	})

	t.Run("cannot add contact without profiles", func(t *testing.T) {
		t.Run("owner must exist", func(t *testing.T) {
			service := NewTestService(t)
			friend := newTestProfile(t, service, "friend")

			_, err := service.AddToContacts(t.Context(), genUserId(), friend.Id)
			if !errors.Is(err, errs.ErrUserNotExists) {
				t.Fatalf("expected ErrUserNotExists for missing owner, got %v", err)
			}
		})

		t.Run("target must exist", func(t *testing.T) {
			service := NewTestService(t)
			owner := newTestProfile(t, service, "owner")

			_, err := service.AddToContacts(t.Context(), owner.Id, genUserId())
			if !errors.Is(err, errs.ErrUserNotExists) {
				t.Fatalf("expected ErrUserNotExists for missing target, got %v", err)
			}
		})
	})

	t.Run("list contacts requires profile", func(t *testing.T) {
		service := NewTestService(t)
		missingId := genUserId()

		_, err := service.ListUserContacts(t.Context(), missingId)
		if !errors.Is(err, errs.ErrUserNotExists) {
			t.Fatalf("expected ErrUserNotExists when listing contacts for missing profile, got %v", err)
		}
	})

	t.Run("cannot add self as contact", func(t *testing.T) {
		service := NewTestService(t)
		user := newTestProfile(t, service, "self")

		_, err := service.AddToContacts(t.Context(), user.Id, user.Id)
		if err == nil {
			t.Fatalf("expected error when adding self as contact, got none")
		}
		if !errors.Is(err, errs.ErrSelfContact) {
			t.Fatalf("expected ErrSelfContact, got %v", err)
		}
	})
}

func TestGroup(t *testing.T) {
}
