package auth

import (
	"context"
	"crypto/rand"
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
	"github.com/patrickmn/go-cache"

	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/elug3/gochat/pkg/auth/internal/store"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	store    store.AuthStore
	jwtKey   *rsa.PrivateKey
	jwks     *jwk.Jwks
	wsTokens *cache.Cache
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
	var jwtKey *rsa.PrivateKey
	var err error

	if opts.UseTmpKey {
		log.Warn().Msg("using temporary RSA key, not recommended for production")
		jwtKey, err = rsa.GenerateKey(rand.Reader, 2048)
	} else {
		jwtKey, err = loadPrivateKey(opts.KeyPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA private key: %w", err)
	}

	// Create JWKs
	jwks := jwk.NewJwks()
	err = jwks.AddKey(jwtKey.PublicKey, "key1")
	if err != nil {
		return nil, fmt.Errorf("failed to add key to JWKs: %w", err)
	}

	// Initialize store
	store, err := sqlite3.NewAuthStore(opts.SaveDir, opts.InMemory)
	if err != nil {
		return nil, err
	}
	return &AuthService{
		store:    store,
		jwtKey:   jwtKey,
		jwks:     jwks,
		wsTokens: cache.New(5*time.Minute, 10*time.Minute),
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
		return 0, fmt.Errorf("%w: %w", errs.ErrConstraint, err)
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

// ValidateAccessToken validates the access token and returns the user Id if valid.
func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (userId int32, err error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtKey.Public(), nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, errs.ErrAuthenticationFailure
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(string); ok {
			var uid int32
			if _, err = fmt.Sscan(sub, &uid); err != nil {
				return 0, fmt.Errorf("invalid sub claim: %w", err)
			}
			return uid, nil
		}
		return 0, fmt.Errorf("sub claim not found")
	}
	return 0, fmt.Errorf("invalid token claims")
}

func (s *AuthService) RevokeAccessToken(ctx context.Context) error {
	return nil
}

func (s *AuthService) CreateWsToken(ctx context.Context, accessToken string) (string, error) {
	userId, err := s.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return "", err
	}
	return s.createWsToken(ctx, userId)
}

func (s *AuthService) createWsToken(ctx context.Context, userId int32) (string, error) {
	token := rand.Text()
	s.wsTokens.Set(token, userId, time.Second*300) // 5 minutes

	return token, nil
}

func (s *AuthService) UseWsToken(ctx context.Context, token string) (userId int32, err error) {
	u, found := s.wsTokens.Get(token)
	if !found {
		return 0, errs.ErrAuthenticationFailure
	}
	userId, ok := u.(int32)
	if !ok {
		return 0, fmt.Errorf("invalid user id type in ws token cache")
	}
	s.wsTokens.Delete(token)
	return userId, nil
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
