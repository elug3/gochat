package service

import (
	"context"
	"fmt"
	"time"

	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
)

type UserService struct {
	store store.UserStore
}

func NewUserService(userStore store.UserStore) (*UserService, error) {
	s := UserService{store: userStore}
	return &s, nil
}

func (s *UserService) GetUser(ctx context.Context, userId int) (*model.User, error) {
	txu, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txu.Rollback()
	user, err := txu.GetUser(userId)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}
	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, tokenString string) (*model.Token, error) {
	txu, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txu.Rollback()
	token, err := txu.GetToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("GetToken: %w", err)
	}
	return token, nil
}

func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	txu, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txu.Rollback()

	user, err := txu.CreateUser(username)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	if err = txu.UpdatePassword(user.Id, password); err != nil {
		return nil, err
	}
	if err = txu.Commit(); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*model.Token, error) {
	txu, err := s.store.Begin()
	if err != nil {
		return nil, err
	}
	defer txu.Rollback()

	userId, err := txu.ValidatePassword(username, password)
	if err != nil {
		return nil, err
	}
	token, err := txu.CreateToken(userId, time.Hour*24*7)
	if err != nil {
		return nil, err
	}
	if err = txu.Commit(); err != nil {
		return nil, err
	}
	return token, nil
}
