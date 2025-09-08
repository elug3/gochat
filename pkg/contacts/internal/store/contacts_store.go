package store

import (
	"github.com/elug3/gochat/internal/services/contcts_service/access"
	"github.com/elug3/gochat/internal/services/contcts_service/model"
)

type ContactsStore interface {
	Begin() (TxContacts, error)
}

type TxContacts interface {
	Rollback() error
	Commit() error

	GetGroup(id int) (*model.Group, error)
	// ListGroups returns a list of groups
	ListGroups(limit int) ([]model.Group, error)
	CreateGroup(name string) (*model.Group, error)
	DeleteGroup(id int) error

	ListUserGroups(userId int) ([]model.Group, error)
	ListGroupMembers(groupId int) ([]model.Member, error)

	GetMember(groupId, userId int) (*model.Member, error)
	CreateMember(groupId, userId int, role access.Role) (*model.Member, error)
	DeleteMember(groupId, userId int) error
	MemberExists(groupId, userId int) (bool, error)

	ListProfiles(limit int) ([]model.Profile, error)
	CreateProfile(userId int, name string) (*model.Profile, error)
	DeleteProfile(userId int) error

	// FindOwners returns a list of members with RoleOwner for the given user
	FindOwners(userId int) ([]model.Member, error)
}
