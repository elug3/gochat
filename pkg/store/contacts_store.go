package store

import "github.com/elug3/gochat/pkg/model"

type ContactsStore interface {
	Begin() (TxContacts, error)
}

type TxcKey struct{}

type TxContacts interface {
	Rollback() error
	Commit() error

	GetGroups(userId int) ([]model.Group, error)
	GetGroup(id int) (*model.Group, error)
	CreateGroup(name string) (*model.Group, error)
	DeleteGroup(id int) error

	GetMembers(groupId int) ([]model.Member, error)
	GetMember(groupId, userId int) (*model.Member, error)
	CreateMember(groupId, userId int, role model.Role) (*model.Member, error)
	DeleteMember(groupId, userId int) error
	MemberExists(groupId, userId int) (bool, error)

	CreateProfile(userId int, name string) (*model.Profile, error)
	DeleteProfile(userId int) error
}
