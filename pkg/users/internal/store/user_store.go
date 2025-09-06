package store

import (
	"context"

	"github.com/elug3/gochat/pkg/users/internal/model"
)

type UserStore interface {
	Begin() (TxUser, error)
	BeginTx(context.Context) (TxUser, error)
}

type TxUser interface {
	Rollback() error
	Commit() error

	GetUser(userId int32) (*model.User, error)
	CreateUser(username string) (*model.User, error)
	UpdateUser(userId int32, username string) (*model.User, error)
	DeleteUser(userId int32) error
}
