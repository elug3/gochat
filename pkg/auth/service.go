package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/jwk"
	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/elug3/gochat/pkg/auth/internal/store"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	store  store.AuthStore
	jwtKey *rsa.PrivateKey
	jwks   *jwk.Jwks
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	der, _ := pem.Decode(keyData)
	key, err := x509.ParsePKCS8PrivateKey(der.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return rsaKey, nil
}

func NewAuthService(opts *Options) (*AuthService, error) {
	// Load private key
	jwtKey, err := loadPrivateKey(opts.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Create JWKs
	jwks := jwk.NewJwks()
	err = jwks.AddKey(jwtKey.PublicKey, "key1")
	if err != nil {
		return nil, fmt.Errorf("failed to add key to JWKs: %w", err)
	}

	// Initialize store
	store, err := sqlite3.NewAuthStore()
	if err != nil {
		return nil, err
	}
	return &AuthService{
		store:  store,
		jwtKey: jwtKey,
		jwks:   jwks,
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

func (s *AuthService) RegisterUser(ctx context.Context, username, password string) (int32, error) {
	err := credentialsRule(username, password)
	if err != nil {
		return 0, err
	}
	passwordHash, err := newHash(password)
	if err != nil {
		return 0, err
	}

	userId, err := s.store.CreateCredential(ctx, username, passwordHash)
	if err != nil {
		return 0, fmt.Errorf("cannot create user '%s': %w", username, err)
	}

	return userId, nil
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

func (s *AuthService) GetJwksData() ([]byte, error) {
	if s.jwks == nil {
		return nil, errors.New("jwks not initialized")
	}
	return s.jwks.Json()
}

func (s *AuthService) Close() error {
	return s.store.Close()
}

func (s *AuthService) newClaims(userId int32, expiresIn time.Duration) (accessToken string, err error) {
	issudAt := time.Now()
	expiresAt := issudAt.Add(expiresIn)

	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": strconv.Itoa(int(userId)),
		"iat": jwt.NewNumericDate(issudAt),
		"exp": jwt.NewNumericDate(expiresAt),
		"aud": "api",
	})
	claims.Header["kid"] = "key1"

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
