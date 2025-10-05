package store

import (
	"github.com/elug3/gochat/pkg/contacts/access"
	"github.com/elug3/gochat/pkg/contacts/internal/model"
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

	ListUserGroups(userId int32) ([]model.Group, error)
	ListGroupMembers(groupId int) ([]model.Member, error)

	GetMember(groupId int, userId int32) (*model.Member, error)
	CreateMember(groupId int, userId int32, role access.Role) (*model.Member, error)
	DeleteMember(groupId int, userId int32) error
	MemberExists(groupId int, userId int32) (bool, error)

	ListProfiles(limit int) ([]model.Profile, error)
	GetProfile(userId int32) (*model.Profile, error)
	CreateProfile(userId int32, name, iconUrl string) (*model.Profile, error)
	UpdateProfile(userId int32, name, iconUrl string) (*model.Profile, error)
	DeleteProfile(userId int32) error

	// CreateContact creates a one-way relationship from userId to targetId
	CreateContact(userId, targetId int32, alias *string) (*model.Contact, error)
	// ListUserContacts returns a list of contacts for the given user
	ListUserContacts(userId int32) ([]model.Contact, error)
	// DeleteUserContact delete specific contact for the given user
	DeleteUserContact(userId, targetId int32) error

	// FindOwners returns a list of members with RoleOwner for the given user
	FindOwners(userId int32) ([]model.Member, error)
}
