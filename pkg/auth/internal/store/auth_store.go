package store

import (
	"context"

	"github.com/elug3/gochat/pkg/auth/internal/model"
)

type AuthStore interface {
	GetCredential(ctx context.Context, username string) (*model.Credentials, error)
	SaveCredentials(ctx context.Context, userId int32, username string, hashedPassword string) error
	UpdatePassword(ctx context.Context, userId int32, newHashedPassword string) error
	Close() error
}
