package contacts

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/errs"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
	"github.com/elug3/gochat/pkg/contacts/internal/store"
	"github.com/elug3/gochat/pkg/contacts/internal/store/s3"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
	"github.com/elug3/identicon"
	"github.com/rs/zerolog/log"
)

type ContactsService struct {
	store     store.ContactsStore
	pub       *events.Publisher
	iconStore *s3.IconStore
}

func NewContactsService(opts *Options) (*ContactsService, error) {
	contactsStore, err := sqlite3.NewContactsStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return nil, err
	}
	pub, err := events.NewPublisher(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	iconStore, err := s3.NewIconStore(opts.S3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot create s3 icon store: %w", err)
	}

	s := ContactsService{
		store:     contactsStore,
		iconStore: iconStore,
		pub:       pub,
	}
	return &s, nil
}

func (s *ContactsService) GetGroup(groupId int) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	group, err := txc.GetGroup(groupId)
	if err != nil {
		return nil, fmt.Errorf("cannot get group '%d': %w", groupId, err)
	}
	return group, nil
}

func (s *ContactsService) ListGroups() ([]model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	gs, err := txc.ListGroups(50)
	if err != nil {
		return nil, fmt.Errorf("cannot list groups: %w", err)
	}
	return gs, nil
}

func (s *ContactsService) CreateUserGroup(userId int32, groupName string) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	g, err := txc.CreateGroup(groupName)
	if err != nil {
		return nil, fmt.Errorf("cannot create group '%s': %w", groupName, err)

	}
	if _, err = s.join(txc, g.Id, userId, access.RoleOwner); err != nil {
		return nil, fmt.Errorf("cannot add user '%d' to group '%d': %w", userId, g.Id, err)
	}
	if err = txc.Commit(); err != nil {
		return nil, err
	}

	if err = s.pub.Publish(events.GroupCreated{
		GroupId:   g.Id,
		GroupName: g.Name,
		TimeStamp: g.CreatedAt.Unix(),
	}); err != nil {
		log.Err(err).Msgf("cannot publish GroupCreated event for group '%d'", g.Id)
	}
	if err = s.pub.Publish(events.MemberJoined{
		GroupId:   g.Id,
		UserId:    userId,
		TimeStamp: g.CreatedAt.Unix(),
	}); err != nil {
		log.Err(err).Msgf("cannot publish MemberJoined event for user '%d' in group '%d'", userId, g.Id)
	}

	return g, nil
}

func (s *ContactsService) ListUserGroups(userId int32) ([]model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()
	groups, err := txc.ListUserGroups(userId)
	if err != nil {
		return nil, fmt.Errorf("cannot list groups for user '%d': %w", userId, err)
	}
	return groups, nil
}

func (s *ContactsService) GetUserGroup(groupId int, userId int32) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("cannot get group '%d' for user '%d': %w", groupId, userId, err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	group, err := txc.GetGroup(groupId)
	if err != nil {
		return nil, fmt.Errorf("cannot get group '%d': %w", groupId, err)
	}
	return group, nil
}

func (s *ContactsService) DeleteUserGroup(groupId int, userId int32) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

	actMbr, err := txc.GetMember(groupId, userId)
	if err != nil {
		return fmt.Errorf("cannot get member for user '%d' in group '%d': %w", userId, groupId, err)
	}
	if !access.Can(actMbr.Role, actMbr.Role, access.ActionDeleteGroup) {
		return fmt.Errorf("user '%d' cannot delete group '%d': %w", userId, groupId, errs.ErrPermissionDenied)
	}

	if err = s.deleteGroup(txc, groupId); err != nil {
		return fmt.Errorf("cannot delete group '%d': %w", groupId, err)
	}
	if err = txc.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) deleteGroup(txc store.TxContacts, id int) error {
	err := txc.DeleteGroup(id)
	if err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) Invite(groupId int, inviterId int32, inviteeId int32) (*model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Commit()

	if ok, err := s.CanInvite(groupId, inviterId); err != nil || !ok {
		if err != nil {
			return nil, fmt.Errorf("failed to check permission: %w", err)
		}
		// permission denied
		return nil, fmt.Errorf("user '%d' cannot invite members to group '%d': %w", inviterId, groupId, errs.ErrPermissionDenied)
	}
	member, err := s.join(txc, groupId, inviteeId, access.RoleMember)
	if err != nil {
		return nil, fmt.Errorf("cannot join user '%d' to group '%d': %w", inviteeId, groupId, err)
	}

	if err = txc.Commit(); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *ContactsService) join(txc store.TxContacts, groupId int, userId int32, role access.Role) (*model.Member, error) {
	return txc.CreateMember(groupId, userId, role)
}

func (s *ContactsService) ListGroupMembers(groupId int, userId int32) ([]model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("MemberExists: %w", err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	members, err := txc.ListGroupMembers(groupId)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *ContactsService) GetMember(groupId int, userId int32, targetId int32) (*model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("MemberExists: %w", err)
		}
		return nil, fmt.Errorf("user '%d' does not exist in group '%d'", userId, groupId)
	}

	m, err := txc.GetMember(groupId, targetId)
	if err != nil {
		return nil, fmt.Errorf("GetMember: %w", err)
	}
	return m, nil
}

func (s *ContactsService) DeleteMember(groupId int, userId int32, targetId int32) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

	if userId == targetId {
		// A member can always leave a group by themselves. No need to check permissions. except owner.
		if err := s.leave(txc, groupId, userId); err != nil {
			return err
		}
	} else {
		if ok, err := s.CanDeleteMember(groupId, userId, targetId); err != nil || !ok {
			if err != nil {
				return fmt.Errorf("failed to check permission: %w", err)
			}
			// permission denied
			return fmt.Errorf("cannot delete member '%d': %w", targetId, errs.ErrPermissionDenied)
		}

		if err = s.deleteMember(txc, groupId, targetId); err != nil {
			return fmt.Errorf("user '%d' cannot delete member for group '%d', '%d': %w", userId, groupId, targetId, err)
		}
	}
	if err = txc.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) leave(txc store.TxContacts, groupId int, userId int32) error {
	if ok, err := s.CanLeave(groupId, userId); !ok {
		if err != nil {
			return fmt.Errorf("failed to check permission: %w", err)
		}
		return fmt.Errorf("cannot leave group '%d': %w", groupId, errs.ErrPermissionDenied)
	}

	if err := txc.DeleteMember(groupId, userId); err != nil {
		return fmt.Errorf("user '%d' cannot leave group '%d': %w", userId, groupId, err)
	}
	return nil
}

func (s *ContactsService) deleteMember(txc store.TxContacts, groupId int, userId int32) error {
	if err := txc.DeleteMember(groupId, userId); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) ListProfiles(limit int) ([]model.Profile, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	profiles, err := txc.ListProfiles(limit)
	if err != nil {
		return nil, fmt.Errorf("cannot list profiles: %w", err)
	}
	return profiles, nil
}

func (s *ContactsService) genNewIcon(ctx context.Context, userId int32) (url string, err error) {
	idStr := strconv.FormatInt(int64(userId), 10)
	img := identicon.New([]byte(idStr), 256)
	err = s.iconStore.UploadIcon(ctx, idStr, img)
	if err != nil {
		return "", err
	}
	url = fmt.Sprintf("http://localhost:9000/profile-icons/%s.png", idStr)
	return url, nil
}

func (s *ContactsService) CreateProfile(ctx context.Context, userId int32, name string) (*model.Profile, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	iconUrl, err := s.genNewIcon(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot generate icon for user '%d': %w", userId, err)
	}

	profile, err := txc.CreateProfile(userId, name, iconUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot create profile for user '%d': %w", userId, err)
	}

	if err = txc.Commit(); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *ContactsService) GetProfile(userId int32) (*model.Profile, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	profile, err := txc.GetProfile(userId)
	if err != nil {
		return nil, fmt.Errorf("cannot get profile for user '%d': %w", userId, err)
	}
	return profile, nil
}

func (s *ContactsService) DeleteProfile(userId int32) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

	ms, err := txc.FindOwners(userId)
	if err != nil {
		return fmt.Errorf("cannot find groups owned by user '%d': %w", userId, err)
	}
	// A user cannot delete their profile if they are owner of any group.
	if len(ms) > 0 {
		return fmt.Errorf("cannot delete profile for user '%d': user is owner of %d groups: %w", userId, len(ms), errs.ErrPermissionDenied)
	}

	if err = txc.DeleteProfile(userId); err != nil {
		return fmt.Errorf("cannot delete profile for user '%d': %w", userId, err)
	}

	if err = txc.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	return nil
}

func (s *ContactsService) ListUserContacts(userId int32) ([]model.Contact, error) {
}

func (s *ContactsService) AddToContacts(userId int32, targetId int32) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

}

func (s *ContactsService) CanInvite(groupId int, inviterId int32) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, err
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, inviterId)
	if err != nil {
		return false, nil
	}

	if access.Can(m.Role, m.Role, access.ActionInvite) {
		return true, nil
	}
	return false, fmt.Errorf("user '%d' cannot invite members to group '%d': %w", inviterId, groupId, errs.ErrPermissionDenied)
}

func (s *ContactsService) CanSend(groupId int, userId int32) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, err
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		if err == errs.ErrNotFound {
			return false, fmt.Errorf("user is not member of group: %w", errs.ErrPermissionDenied)
		}
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionSend) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanRead(groupId int, userId int32) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, err
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, err
	}
	if access.Can(m.Role, m.Role, access.ActionRead) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanLeave(groupId int, userId int32) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, err
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionLeave) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanDeleteMember(groupId int, userId int32, targetId int32) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, err
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, err
	}

	if access.Can(m.Role, m.Role, access.ActionDeleteMember) {
		return true, nil
	}
	return false, nil
}

type AccessRequest struct {
	UserId   int32
	ChatId   int
	Action   access.Action
	TargetId int32 // optional, used for actions that target another user
}

func (s *ContactsService) Can(req AccessRequest) (bool, access.Action, error) {
	var (
		can bool
		act access.Action
		err error
	)
	act = req.Action

	switch req.Action {
	case access.ActionInvite:
		can, err = s.CanInvite(req.ChatId, req.TargetId)
	case access.ActionSend:
		can, err = s.CanSend(req.ChatId, req.UserId)
	case access.ActionRead:
		can, err = s.CanRead(req.ChatId, req.UserId)
	case access.ActionLeave:
		can, err = s.CanLeave(req.ChatId, req.UserId)
	case access.ActionDeleteMember:
		can, err = s.CanDeleteMember(req.ChatId, req.UserId, req.TargetId)
	default:
		can = false
		act = ""
		err = fmt.Errorf("unknown action: %q", req.Action)
	}
	return can, act, err
}
