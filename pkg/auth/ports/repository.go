package ports

import (
	"context"

	"github.com/elug3/gochat/pkg/auth/domain"
)

// UserRepository defines persistence operations required by auth services.
type UserRepository interface {
	CreateUser(
		ctx context.Context,
		username string,
		passwordHash string,
		name string,
	) (*domain.User, error)

	GetUserByUsername(
		ctx context.Context,
		username string,
	) (*domain.User, error)

	GetUserByID(
		ctx context.Context,
		userID int32,
	) (*domain.User, error)

	UpdatePasswordHash(
		ctx context.Context,
		userID int32,
		passwordHash string,
	) error

	DeleteUser(
		ctx context.Context,
		userID int32,
	) error
}

// SessionRepository defines session persistence operations.
type SessionRepository interface {
	CreateSession(
		ctx context.Context,
		userID int32,
		audiences []string,
		scopes []string,
	) (*domain.Session, error)

	GetSessionByID(
		ctx context.Context,
		sessionID string,
	) (*domain.Session, error)

	ListUserSessions(
		ctx context.Context,
		userID int32,
	) ([]*domain.Session, error)

	RevokeSession(
		ctx context.Context,
		sessionID string,
	) error

	DeleteExpiredSessions(
		ctx context.Context,
	) error
}
