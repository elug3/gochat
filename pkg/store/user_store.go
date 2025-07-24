package store

import (
	"time"

	"github.com/elug3/gochat/pkg/model"
)

type UserStore interface {
	Begin() (TxUser, error)
}

type TxuKey struct{}

type TxUser interface {
	Rollback() error
	Commit() error

	GetUser(userId int) (*model.User, error)
	CreateUser(username string) (*model.User, error)
	UpdatePassword(userId int, password string) error
	ValidatePassword(username, password string) (userId int, err error)
	CreateToken(userId int, expiresIn time.Duration) (*model.Token, error)
	GetToken(tokenString string) (*model.Token, error)
}
