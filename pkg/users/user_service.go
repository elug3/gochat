package users

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/pkg/users/internal/model"
	"github.com/elug3/gochat/pkg/users/internal/store"
)

type UserService struct {
	store store.UserStore
}

type ServiceDeps struct {
	Store store.UserStore
}

func NewUserService(deps ServiceDeps) (*UserService, error) {
	if deps.Store == nil {
		return nil, fmt.Errorf("user store is required")
	}
	return &UserService{
		store: deps.Store,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userId int32) (*model.User, error) {
	txu, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txu.Rollback()
	user, err := s.getUser(txu, userId)
	if err != nil {
		return nil, fmt.Errorf("cannot get user: %d: %w", userId, err)
	}
	return user, nil
}

func (s *UserService) getUser(txu store.TxUser, userId int32) (*model.User, error) {
	user, err := txu.GetUser(userId)
	return user, err
}

// func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {

// }
