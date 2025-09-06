package service

import (
	"fmt"
	"time"

	"github.com/elug3/gochat/internal/services/contcts_service/access"
	"github.com/elug3/gochat/internal/services/contcts_service/model"
	"github.com/elug3/gochat/internal/services/contcts_service/store"
	"github.com/elug3/gochat/pkg/events"
)

type ContactsService struct {
	store store.ContactsStore
	ep    *events.EventPublisher
}

func NewContactsService(contactsStore store.ContactsStore, ep *events.EventPublisher) (*ContactsService, error) {
	s := ContactsService{
		store: contactsStore,
		ep:    ep,
	}
	return &s, nil
}

func (s *ContactsService) CreateGroup(userId int, groupName string) (*model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	group, err := txc.CreateGroup(groupName)
	if err != nil {
		return nil, fmt.Errorf("CreateGroup: %w", err)
	}
	if _, err = s.join(txc, group.Id, userId, access.RoleOwner); err != nil {
		return nil, fmt.Errorf("CreateMember: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return nil, fmt.Errorf("Commit: %w", err)
	}

	if err = s.ep.Publish(events.GroupCreated{
		GroupId:   group.Id,
		GroupName: group.Name,
		TimeStamp: time.Now().Unix(),
	}); err != nil {
		return nil, fmt.Errorf("Publish: %w", err)
	}
	if err = s.ep.Publish(events.MemberJoined{
		GroupId:   group.Id,
		UserId:    userId,
		Role:      access.RoleOwner,
		TimeStamp: time.Now().Unix(),
	}); err != nil {
		return nil, fmt.Errorf("Publish: %w", err)
	}

	return group, nil
}

func (s *ContactsService) ListGroups(userId int) ([]model.Group, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
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
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	exists, err := txc.MemberExists(groupId, userId)
	if err != nil {
		return nil, fmt.Errorf("MemberExists: %w", err)
	}
	if exists {
		group, err := txc.GetGroup(groupId)
		if err != nil {
			return nil, fmt.Errorf("GetGroup: %w", err)
		}
		return group, nil
	}
	// return nil, &store.Error{
	// 	Kind:    store.KindGroup,
	// 	Err:     store.ErrNotFound,
	// 	Message: fmt.Sprintf("cannot find group %d for user %d", groupId, userId),
	// }
}

func (s *ContactsService) DeleteGroup(groupId, userId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	actMbr, err := txc.GetMember(groupId, userId)
	if err != nil {
		return fmt.Errorf("GetMember: %w", err)
	}
	if !access.Can(actMbr.Role, actMbr.Role, access.ActionDeleteGroup) {
		// return &store.Error{
		// 	Kind:    store.KindMember,
		// 	Err:     store.ErrPermissionDenied,
		// 	Message: "permission denided",
		// }
	}

	if err = s.deleteGroup(txc, groupId); err != nil {
		return fmt.Errorf("DeleteGroup: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

func (s *ContactsService) deleteGroup(txc store.TxContacts, id int) error {
	err := txc.DeleteGroup(id)
	if err != nil {
		return fmt.Errorf("DeleteGroup: %w", err)
	}
	return nil
}

func (s *ContactsService) Invite(groupId, inviterId, inviteeId int) (*model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Commit()

	if !s.CanInvite(txc, groupId, inviterId) {
		// return nil, &store.Error{
		// 	Kind:    store.KindMember,
		// 	Err:     store.ErrPermissionDenied,
		// 	Message: "permission denied",
		// }
	}
	member, err := s.join(txc, groupId, inviteeId, access.RoleMember)
	if err != nil {
		return nil, fmt.Errorf("CreateMember: %w", err)
	}

	if err = txc.Commit(); err != nil {
		return nil, fmt.Errorf("Commit: %w", err)
	}
	return member, nil
}

func (s *ContactsService) join(txc store.TxContacts, groupId, userId int, role access.Role) (*model.Member, error) {
	return txc.CreateMember(groupId, userId, role)
}

func (s *ContactsService) ListMembers(groupId, userId int) ([]model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("MemberExists: %w", err)
		}
		// return nil, &store.Error{
		// 	Kind:    store.KindMember,
		// 	Err:     store.ErrNotFound,
		// 	Message: fmt.Sprintf("user '%d' does not exist in group '%d'", userId, groupId),
		// }
	}

	members, err := txc.GetMembers(groupId)
	if err != nil {
		return nil, fmt.Errorf("GetMembers: %w", err)
	}
	return members, nil
}

func (s *ContactsService) GetMember(groupId, userId, targetId int) (*model.Member, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	if exists, err := txc.MemberExists(groupId, userId); !exists {
		if err != nil {
			return nil, fmt.Errorf("MemberExists: %w", err)
		}
		// return nil, &store.Error{
		// 	Kind:    store.KindMember,
		// 	Err:     store.ErrNotFound,
		// 	Message: fmt.Sprintf("user '%d' does not exist in group '%d'", userId, groupId),
		// }
	}

	m, err := txc.GetMember(groupId, targetId)
	if err != nil {
		return nil, fmt.Errorf("GetMember: %w", err)
	}
	return m, nil
}

func (s *ContactsService) DeleteMember(groupId, userId, targetId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	if userId == targetId {
		if err := s.leave(txc, groupId, userId); err != nil {
			return fmt.Errorf("Leave: %w", err)
		}
	} else {
		if ok, err := s.canDeleteMember(txc, groupId, userId, targetId); !ok {
			if err != nil {
				return fmt.Errorf("CanDeleteMember: %w", err)
			}
			// return &store.Error{
			// 	Kind:    store.KindMember,
			// 	Err:     store.ErrPermissionDenied,
			// 	Message: "permission denied",
			// }
		}
		if err = s.deleteMember(txc, groupId, targetId); err != nil {
			return fmt.Errorf("DeleteMember: %w", err)
		}
	}
	if err = txc.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

func (s *ContactsService) leave(txc store.TxContacts, groupId, userId int) error {
	if ok, err := s.canLeave(txc, groupId, userId); !ok {
		if err != nil {
			return fmt.Errorf("CanLeave: %w", err)
		}
		// return &store.Error{
		// 	Kind:    store.KindMember,
		// 	Err:     store.ErrPermissionDenied,
		// 	Message: "permission denied",
		// }
	}

	if err := txc.DeleteMember(groupId, userId); err != nil {
		return fmt.Errorf("DeleteMember: %w", err)
	}
	return nil
}

func (s *ContactsService) deleteMember(txc store.TxContacts, groupId, userId int) error {
	if err := txc.DeleteMember(groupId, userId); err != nil {
		return fmt.Errorf("DeleteMember: %w", err)
	}
	return nil
}

func (s *ContactsService) CreateProfile(userId int, name string) (*model.Profile, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Commit()

	profile, err := txc.CreateProfile(userId, name)
	if err != nil {
		return nil, fmt.Errorf("CreateProfile: %w", err)
	}
	if err = txc.Commit(); err != nil {
		return nil, fmt.Errorf("Commit: %w", err)
	}

	return profile, nil
}

func (s *ContactsService) DeleteProfile(userId int) error {
	txc, err := s.store.Begin()
	if err != nil {
		return fmt.Errorf("Begin: %w", err)
	}
	defer txc.Commit()

	if err = txc.DeleteProfile(userId); err != nil {
		return fmt.Errorf("DeleteProfile: %w", err)
	}
	return nil
}

func (s *ContactsService) CanInvite(txc store.TxContacts, groupId, inviterId int) bool {
	m, err := txc.GetMember(groupId, inviterId)
	if err != nil {
		return false
	}

	if access.Can(m.Role, m.Role, access.ActionInvite) {
		return true
	}
	return false
}

func (s *ContactsService) CanSend(groupId, userId int) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()

	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, fmt.Errorf("GetMember: %w", err)
	}

	if access.Can(m.Role, m.Role, access.ActionSend) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) CanRead(groupId, userId int) (bool, error) {
	txc, err := s.store.Begin()
	if err != nil {
		return false, fmt.Errorf("Begin: %w", err)
	}
	defer txc.Rollback()
	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, fmt.Errorf("GetMember: %w", err)
	}
	if access.Can(m.Role, m.Role, access.ActionRead) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) canLeave(txc store.TxContacts, groupId, userId int) (bool, error) {
	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, fmt.Errorf("GetMember: %w", err)
	}

	if access.Can(m.Role, m.Role, access.ActionLeave) {
		return true, nil
	}
	return false, nil
}

func (s *ContactsService) canDeleteMember(txc store.TxContacts, groupId, userId, targetId int) (bool, error) {
	m, err := txc.GetMember(groupId, userId)
	if err != nil {
		return false, fmt.Errorf("GetMember: %w", err)
	}

	if access.Can(m.Role, m.Role, access.ActionDeleteMember) {
		return true, nil
	}
	return false, nil
}
