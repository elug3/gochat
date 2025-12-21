package contacts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/elug3/gochat/pkg/contacts/internal/store/s3"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

const (
	// DefaultListLimit is the default limit for list operations
	DefaultListLimit = 50
	// MaxListLimit is the maximum allowed limit for list operations
	MaxListLimit = 1000
)

type ContactsService struct {
	store       *sqlite3.ContactsStore
	pub         *events.Publisher
	iconStore   *s3.IconStore
	iconBaseURL string
}

var contactsResetOnce sync.Once

func NewContactsService(opts *Options) (*ContactsService, error) {
	var (
		pub       *events.Publisher
		iconStore *s3.IconStore
	)
	contactsStore, isNewStore, err := sqlite3.NewContactsStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return nil, err
	}

	if !opts.NoEvent {
		pub, err = events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return nil, fmt.Errorf("cannot create event publisher: %w", err)
		}
	}

	if !opts.NoIcons {
		iconStore, err = s3.NewIconStore(opts.S3Endpoint, opts.S3AccessKey, opts.S3SecretKey, opts.S3Region)
		if err != nil {
			return nil, fmt.Errorf("cannot create s3 icon store: %w", err)
		}
	}

	s := ContactsService{
		store:       contactsStore,
		iconStore:   iconStore,
		pub:         pub,
		iconBaseURL: opts.IconBaseURL,
	}

	// if isNewStore, publish reset event
	if isNewStore {
		log.Info().Msg("new contacts store created")
		s.publishResetEvent()
	}
	return &s, nil
}

func (s *ContactsService) publish(event events.Event, errMsg string, args ...interface{}) {
	if s.pub == nil {
		return
	}

	if err := s.pub.Publish(event); err != nil {
		log.Err(err).Msgf(errMsg, args...)
	}
}

func (s *ContactsService) publishResetEvent() {
	if s.pub == nil {
		return
	}

	contactsResetOnce.Do(func() {
		s.publish(events.ContactsReset{
			TimeStamp: time.Now().Unix(),
		}, "cannot publish ContactsReset event after initialization")
	})
}

func (s *ContactsService) GetGroup(ctx context.Context, groupId int) (*model.Group, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	group, err := s.store.GetGroup(tx, groupId)
	if err != nil {
		return nil, fmt.Errorf("cannot get group '%d': %w", groupId, err)
	}
	return group, nil
}

func (s *ContactsService) ListGroups(ctx context.Context, limit int) ([]model.Group, error) {
	if limit <= 0 || limit > MaxListLimit {
		limit = DefaultListLimit
	}

	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	gs, err := s.store.ListGroups(tx, limit)
	if err != nil {
		return nil, fmt.Errorf("cannot list groups: %w", err)
	}
	return gs, nil
}

func (s *ContactsService) CreateUserGroup(ctx context.Context, userId int32, groupName string) (*model.Group, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	g, err := s.store.CreateGroup(tx, groupName)
	if err != nil {
		return nil, fmt.Errorf("cannot create group '%s': %w", groupName, err)

	}
	if _, err = s.join(tx, g.Id, userId, access.RoleOwner); err != nil {
		return nil, fmt.Errorf("cannot add user '%d' to group '%d': %w", userId, g.Id, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("cannot commit transaction: %w", err)
	}

	s.publish(events.GroupCreated{
		GroupId:   g.Id,
		GroupName: g.Name,
		TimeStamp: g.CreatedAt.Unix(),
	}, "cannot publish GroupCreated event for group '%d'", g.Id)

	s.publish(events.MemberJoined{
		GroupId:   g.Id,
		UserId:    userId,
		TimeStamp: g.CreatedAt.Unix(),
	}, "cannot publish MemberJoined event for user '%d' in group '%d'", userId, g.Id)

	return g, nil
}

func (s *ContactsService) ListUserGroups(ctx context.Context, userId int32) ([]model.Group, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	groups, err := s.store.ListUserGroups(tx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot list groups for user '%d': %w", userId, err)
	}
	return groups, nil
}

func (s *ContactsService) GetUserGroup(ctx context.Context, groupId int, userId int32) (*model.Group, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	if exists, err := s.store.MemberExists(tx, groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("cannot get group '%d' for user '%d': %w", groupId, userId, err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	group, err := s.store.GetGroup(tx, groupId)
	if err != nil {
		return nil, fmt.Errorf("cannot get group '%d': %w", groupId, err)
	}
	return group, nil
}

func (s *ContactsService) DeleteUserGroup(ctx context.Context, groupId int, userId int32) error {
	tx, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	actMbr, err := s.store.GetMember(tx, groupId, userId)
	if err != nil {
		return fmt.Errorf("cannot get member for user '%d' in group '%d': %w", userId, groupId, err)
	}
	if !access.Can(actMbr.Role, actMbr.Role, access.ActionDeleteGroup) {
		return fmt.Errorf("user '%d' cannot delete group '%d': %w", userId, groupId, errs.ErrPermissionDenied)
	}

	if err = s.deleteGroup(tx, groupId); err != nil {
		return fmt.Errorf("cannot delete group '%d': %w", groupId, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}

	s.publish(events.GroupDeleted{
		GroupId:   groupId,
		TimeStamp: time.Now().Unix(),
	}, "cannot publish GroupDeleted event for group '%d'", groupId)

	return nil
}

func (s *ContactsService) deleteGroup(tx *sql.Tx, id int) error {
	err := s.store.DeleteGroup(tx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) Invite(ctx context.Context, groupId int, inviterId int32, inviteeId int32) (*model.Member, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	if ok, err := s.CanInvite(ctx, groupId, inviterId); err != nil || !ok {
		if err != nil {
			return nil, fmt.Errorf("failed to check permission: %w", err)
		}
		// permission denied
		return nil, fmt.Errorf("user '%d' cannot invite members to group '%d': %w", inviterId, groupId, errs.ErrPermissionDenied)
	}
	member, err := s.join(tx, groupId, inviteeId, access.RoleMember)
	if err != nil {
		return nil, fmt.Errorf("cannot join user '%d' to group '%d': %w", inviteeId, groupId, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("cannot commit transaction: %w", err)
	}

	s.publish(events.MemberJoined{
		GroupId:   groupId,
		UserId:    inviteeId,
		TimeStamp: member.CreatedAt.Unix(),
	}, "cannot publish MemberJoined event for user '%d' in group '%d'", inviteeId, groupId)

	return member, nil
}

func (s *ContactsService) join(tx *sql.Tx, groupId int, userId int32, role access.Role) (*model.Member, error) {
	return s.store.CreateMember(tx, groupId, userId, role)
}

func (s *ContactsService) ListGroupMembers(ctx context.Context, groupId int, userId int32) ([]model.Member, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	if exists, err := s.store.MemberExists(tx, groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("cannot check membership: %w", err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	members, err := s.store.ListGroupMembers(tx, groupId)
	if err != nil {
		return nil, fmt.Errorf("cannot list group members: %w", err)
	}
	return members, nil
}

func (s *ContactsService) GetMember(ctx context.Context, groupId int, userId int32, targetId int32) (*model.Member, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	if exists, err := s.store.MemberExists(tx, groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("cannot check membership: %w", err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	m, err := s.store.GetMember(tx, groupId, targetId)
	if err != nil {
		return nil, fmt.Errorf("cannot get member: %w", err)
	}
	return m, nil
}

func (s *ContactsService) DeleteMember(ctx context.Context, groupId int, userId int32, targetId int32) error {
	tx, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	if userId == targetId {
		// A member can always leave a group by themselves. No need to check permissions. except owner.
		if err := s.leave(ctx, tx, groupId, userId); err != nil {
			return err
		}
	} else {
		if ok, err := s.CanDeleteMember(ctx, groupId, userId, targetId); err != nil || !ok {
			if err != nil {
				return fmt.Errorf("failed to check permission: %w", err)
			}
			// permission denied
			return fmt.Errorf("cannot delete member '%d': %w", targetId, errs.ErrPermissionDenied)
		}

		if err = s.deleteMember(tx, groupId, targetId); err != nil {
			return fmt.Errorf("user '%d' cannot delete member for group '%d', '%d': %w", userId, groupId, targetId, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}

	s.publish(events.MemberLeft{
		GroupId:   groupId,
		UserId:    targetId,
		TimeStamp: time.Now().Unix(),
	}, "cannot publish MemberLeft event for user '%d' in group '%d'", targetId, groupId)

	return nil
}

func (s *ContactsService) leave(ctx context.Context, tx *sql.Tx, groupId int, userId int32) error {
	if ok, err := s.CanLeave(ctx, groupId, userId); !ok {
		if err != nil {
			return fmt.Errorf("failed to check permission: %w", err)
		}
		return fmt.Errorf("cannot leave group '%d': %w", groupId, errs.ErrPermissionDenied)
	}

	if err := s.store.DeleteMember(tx, groupId, userId); err != nil {
		return fmt.Errorf("user '%d' cannot leave group '%d': %w", userId, groupId, err)
	}
	return nil
}

func (s *ContactsService) deleteMember(tx *sql.Tx, groupId int, userId int32) error {
	if err := s.store.DeleteMember(tx, groupId, userId); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) ListProfiles(ctx context.Context, limit int) ([]model.Profile, error) {
	if limit <= 0 || limit > MaxListLimit {
		limit = DefaultListLimit
	}

	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	profiles, err := s.store.ListProfiles(tx, limit)
	if err != nil {
		return nil, fmt.Errorf("cannot list profiles: %w", err)
	}
	return profiles, nil
}

func (s *ContactsService) CreateProfile(ctx context.Context, userId int32, name string) (*model.Profile, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	profile, err := s.store.CreateProfile(tx, userId, name)
	if err != nil {
		return nil, fmt.Errorf("cannot create profile for user '%d': %w", userId, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("cannot commit transaction: %w", err)
	}

	s.publish(events.ProfileCreated{
		UserId:    userId,
		TimeStamp: time.Now().Unix(),
	}, "cannot publish ProfileCreated event for user '%d'", userId)

	return profile, nil
}

func (s *ContactsService) GetProfile(ctx context.Context, userId int32) (*model.Profile, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	profile, err := s.store.GetProfile(tx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot get profile for user '%d': %w", userId, err)
	}
	return profile, nil
}

func (s *ContactsService) DeleteProfile(ctx context.Context, userId int32) error {
	tx, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	ms, err := s.store.FindOwners(tx, userId)
	if err != nil {
		return fmt.Errorf("cannot find groups owned by user '%d': %w", userId, err)
	}
	// A user cannot delete their profile if they are owner of any group.
	if len(ms) > 0 {
		return fmt.Errorf("cannot delete profile for user '%d': user is owner of %d groups: %w", userId, len(ms), errs.ErrPermissionDenied)
	}

	if err = s.store.DeleteProfile(tx, userId); err != nil {
		return fmt.Errorf("cannot delete profile for user '%d': %w", userId, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}

	return nil
}

func (s *ContactsService) ListUserContacts(ctx context.Context, userId int32) ([]model.Contact, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	contacts, err := s.store.ListUserContacts(tx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot list contacts for user '%d': %w", userId, err)
	}
	return contacts, nil
}

func (s *ContactsService) AddToContacts(ctx context.Context, ownerId int32, targetId int32) (*model.Contact, error) {
	if ownerId == targetId {
		return nil, fmt.Errorf("cannot add self to contacts: %w", errs.ErrSelfContact)
	}

	tx, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()

	contact, err := s.store.CreateContact(tx, ownerId, targetId, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot add user '%d' to contacts: %w", targetId, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("cannot commit transaction: %w", err)
	}
	return contact, nil
}

func (s *ContactsService) CanInvite(ctx context.Context, groupId int, inviterId int32) (bool, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	m, err := s.store.GetMember(tx, groupId, inviterId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("user '%d' is not a member of group '%d': %w", inviterId, groupId, errs.ErrPermissionDenied)
		}
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionInvite) {
		return true, nil
	}
	return false, fmt.Errorf("user '%d' cannot invite members to group '%d': %w", inviterId, groupId, errs.ErrPermissionDenied)
}

func (s *ContactsService) CanSend(ctx context.Context, groupId int, userId int32) (bool, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	m, err := s.store.GetMember(tx, groupId, userId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("user is not member of group: %w", errs.ErrPermissionDenied)
		}
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionSend) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanRead(ctx context.Context, groupId int, userId int32) (bool, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	m, err := s.store.GetMember(tx, groupId, userId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("user '%d' is not a member of group '%d': %w", userId, groupId, errs.ErrPermissionDenied)
		}
		return false, err
	}
	if access.Can(m.Role, m.Role, access.ActionRead) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanLeave(ctx context.Context, groupId int, userId int32) (bool, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	m, err := s.store.GetMember(tx, groupId, userId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("user '%d' is not a member of group '%d': %w", userId, groupId, errs.ErrPermissionDenied)
		}
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionLeave) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanDeleteMember(ctx context.Context, groupId int, userId int32, targetId int32) (bool, error) {
	tx, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	actor, err := s.store.GetMember(tx, groupId, userId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("user '%d' is not a member of group '%d': %w", userId, groupId, errs.ErrPermissionDenied)
		}
		return false, err
	}

	target, err := s.store.GetMember(tx, groupId, targetId)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return false, fmt.Errorf("target user '%d' is not a member of group '%d': %w", targetId, groupId, errs.ErrNotFound)
		}
		return false, err
	}

	if access.Can(actor.Role, target.Role, access.ActionDeleteMember) {
		return true, nil
	}
	return false, fmt.Errorf("user '%d' cannot delete member '%d' in group '%d': %w", userId, targetId, groupId, errs.ErrPermissionDenied)
}

type AccessRequest struct {
	UserId   int32
	ChatId   int
	Action   access.Action
	TargetId int32 // optional, used for actions that target another user
}

func (s *ContactsService) Can(ctx context.Context, req AccessRequest) (bool, access.Action, error) {
	var (
		can bool
		act access.Action
		err error
	)
	act = req.Action

	switch req.Action {
	case access.ActionInvite:
		can, err = s.CanInvite(ctx, req.ChatId, req.UserId)
	case access.ActionSend:
		can, err = s.CanSend(ctx, req.ChatId, req.UserId)
	case access.ActionRead:
		can, err = s.CanRead(ctx, req.ChatId, req.UserId)
	case access.ActionLeave:
		can, err = s.CanLeave(ctx, req.ChatId, req.UserId)
	case access.ActionDeleteMember:
		can, err = s.CanDeleteMember(ctx, req.ChatId, req.UserId, req.TargetId)
	default:
		can = false
		act = ""
		err = fmt.Errorf("unknown action: %q", req.Action)
	}
	return can, act, err
}
