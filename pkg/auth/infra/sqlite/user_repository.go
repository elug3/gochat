package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elug3/gochat/pkg/auth/domain"
	"github.com/elug3/gochat/pkg/auth/ports"
)

var _ ports.UserRepository = (*UserRepository)(nil)

// UserRepository implements auth user persistence using sqlite.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	username string,
	passwordHash string,
	name string,
) (*domain.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*domain.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) GetUserByID(
	ctx context.Context,
	userID int32,
) (*domain.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *UserRepository) UpdatePasswordHash(
	ctx context.Context,
	userID int32,
	passwordHash string,
) error {
	return fmt.Errorf("not implemented")
}

func (r *UserRepository) DeleteUser(
	ctx context.Context,
	userID int32,
) error {
	return fmt.Errorf("not implemented")
}
