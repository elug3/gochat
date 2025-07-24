package service

import (
	"fmt"

	"github.com/elug3/gochat/pkg/access"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
)

type ContactsService struct {
	store  store.ContactsStore
	access access.ContactsAccess
}

func NewContactsService(contactsStore store.ContactsStore) (*ContactsService, error) {
	s := ContactsService{
		store: contactsStore,
	}
	return &s, nil
}

// func (s *ContactsService) GetGroups(userId string) (groups []model.Group, err error) {
// }

// func (s *ContactsService) GetGroup(groupId, userId string) (*model.Group, error) {
// }

func (s *ContactsService) CreateGroup(userId int, groupName string) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	group, err := txc.CreateGroup(groupName)
	if err != nil {
		return nil, fmt.Errorf("createGroup: %w", err)
	}
	if _, err = s.join(txc, group.Id, userId, access.RoleOwner); err != nil {
		return nil, fmt.Errorf("join: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *ContactsService) GetGroups(userId int) ([]model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()
	groups, err := txc.GetGroups(userId)
	if err != nil {
		return nil, fmt.Errorf("GetGroups: %w", err)
	}
	return groups, nil
}

func (s *ContactsService) GetGroup(groupId int, userId int) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	if exists, _ := txc.MemberExists(groupId, userId); exists {
		group, err := txc.GetGroup(groupId)
		if err != nil {
			return nil, fmt.Errorf("GetGroup: %w", err)
		}
		return group, nil
	}
	return nil, &store.Error{
		Kind:    store.KindGroup,
		Err:     store.ErrNotFound,
		Message: fmt.Sprintf("cannot find group %d for user %d", groupId, userId),
	}

}

func (s *ContactsService) DeleteGroup(groupId, userId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

	actMbr, err := txc.GetMember(groupId, userId)
	if err != nil {
		return fmt.Errorf("getGroup: %q", err)
	}
	if !s.access.Can(actMbr.Role, actMbr.Role, access.ActionDeleteGroup) {
		return &store.Error{
			Kind:    store.KindMember,
			Err:     store.ErrPermissionDenied,
			Message: "permission denided",
		}
	}

	if err = s.deleteGroup(txc, groupId); err != nil {
		return fmt.Errorf("deleteGroup: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) deleteGroup(txc store.TxContacts, id int) error {
	err := txc.DeleteGroup(id)
	return err
}

func (s *ContactsService) canInvite(txc store.TxContacts, groupId, inviterId int) bool {
	actMbr, err := txc.GetMember(groupId, inviterId)
	if err != nil {
		return false
	}
	return s.access.Can(actMbr.Role, actMbr.Role, access.ActionInvite)
}

func (s *ContactsService) Invite(groupId, inviterId, inviteeId int) (*model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Commit()

	if !s.canInvite(txc, groupId, inviterId) {
		return nil, &store.Error{
			Kind:    store.KindMember,
			Err:     store.ErrPermissionDenied,
			Message: "persission denided",
		}
	}
	member, err := s.join(txc, groupId, inviteeId, access.RoleMember)
	if err != nil {
		return nil, err
	}

	if err = txc.Commit(); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *ContactsService) join(txc store.TxContacts, groupId, userId int, role model.Role) (*model.Member, error) {
	return txc.CreateMember(groupId, userId, role)
}

func (s *ContactsService) ListMember(groupId int) ([]model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Rollback()

	members, err := txc.GetMembers(groupId)
	if err != nil {
		return nil, fmt.Errorf("GetMembers: %w", err)
	}
	return members, nil
}

func (s *ContactsService) DeleteMember(groupId, userId, targetId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Rollback()

	actMbr, err := txc.GetMember(groupId, userId)
	if err != nil {
		return err
	}
	tgtMbr, err := txc.GetMember(groupId, targetId)
	if err != nil {
		return err
	}
	if ok := s.access.Can(actMbr.Role, tgtMbr.Role, access.ActionDeleteMember); !ok {
		return fmt.Errorf("permission denided")
	}

	if err := s.deleteMember(txc, groupId, userId); err != nil {
		return fmt.Errorf("deleteMember: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *ContactsService) deleteMember(txc store.TxContacts, groupId, userId int) error {
	return txc.DeleteMember(groupId, userId)
}

func (s *ContactsService) CreateProfile(userId int, name string) (*model.Profile, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txc.Commit()

	profile, err := txc.CreateProfile(userId, name)
	if err != nil {
		return nil, err
	}
	if err = txc.Commit(); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *ContactsService) DeleteProfile(userId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return err
	}
	defer txc.Commit()

	if err = txc.DeleteProfile(userId); err != nil {
		return fmt.Errorf("cannot delete profile: %w", err)
	}
	return nil
}

// func (s *ContactsService) Can(userId, targetId, int, action access.Action) (bool, error) {

// }
