package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/elug3/gochat/pkg/auth/internal/store"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	store  store.AuthStore
	jwtKey *rsa.PrivateKey
}

func NewAuthService(jwtKey *rsa.PrivateKey) (*AuthService, error) {
	store, err := sqlite3.NewAuthStore()
	if err != nil {
		return nil, err
	}
	return &AuthService{
		store:  store,
		jwtKey: jwtKey,
	}, nil
}

func credentialsRule(username, password string) error {
	if len(username) < 2 || len(username) > 32 {
		return fmt.Errorf("username must be between 2 and 32 characters")
	}
	if len(password) < 8 || len(password) > 64 {
		return fmt.Errorf("password must be between 8 and 64 characters")
	}
	return nil
}

func (s *AuthService) CreateCredentials(ctx context.Context, userId int32, username, password string) error {
	err := credentialsRule(username, password)
	if err != nil {
		return err
	}

	passwordHash, err := newHash(password)
	if err != nil {
		return err
	}
	if err = s.store.SaveCredentials(ctx, userId, username, passwordHash); err != nil {
		return fmt.Errorf("cannot create credentials: %w", err)
	}

	return nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, userId int32, password string) error {
	passwordHash, err := newHash(password)
	if err != nil {
		return err
	}
	if err = s.store.UpdatePassword(ctx, userId, passwordHash); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.Token, error) {
	c, err := s.store.GetCredential(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrAuthenticationFailure
		}
		return nil, err
	}
	if match, err := argon2id.ComparePasswordAndHash(password, c.PasswordHash); !match {
		if err != nil {
			return nil, err
		}
		return nil, errs.ErrAuthenticationFailure
	}
	// if authentication success
	expiresIn := time.Hour * 24 * 7
	accessToken, err := s.newClaims(c.UserId, expiresIn) // 7 days
	if err != nil {
		return nil, err
	}
	return &model.Token{
		UserId:       c.UserId,
		AccessToken:  accessToken,
		RefreshToken: "", // TODO
		ExpiresIn:    expiresIn,
	}, nil
}

func Refresh(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	return "", "", nil
}

func ValidateAccessToken(ctx context.Context, accessToken string) (userID int64, err error) {

	return 0, nil
}

func (s *AuthService) RevokeAccessToken(ctx context.Context) error {

	return nil

}

func (s *AuthService) Close() error {
	return s.store.Close()
}

func (s *AuthService) newClaims(userId int32, expiresIn time.Duration) (accessToken string, err error) {
	t := time.Now()
	expiresAt := t.Add(expiresIn)

	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(int64(userId), 10),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(t),
	})
	accessToken, err = claims.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	log.Info().Msgf("Generated access token for user %d", userId)

	return accessToken, nil
}

func newHash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
