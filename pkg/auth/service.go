package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/jwk"
	"github.com/elug3/gochat/shared/events"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/patrickmn/go-cache"

	"github.com/elug3/gochat/pkg/auth/internal/model"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	store *sqlite3.AuthStore
	db    *sql.DB

	jwtKey   *rsa.PrivateKey
	jwks     *jwk.Jwks
	wsTokens *cache.Cache

	wAuth *webauthn.WebAuthn

	eventPub *events.Publisher
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
	var (
		jwtKey   *rsa.PrivateKey
		eventPub *events.Publisher
		err      error
	)

	// load private key
	if opts.UseTmpKey {
		log.Warn().Msg("using temporary RSA key, not recommended for production")
		jwtKey, err = rsa.GenerateKey(rand.Reader, 2048)
	} else {
		jwtKey, err = loadPrivateKey(opts.KeyPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load RSA private key: %w", err)
	}

	// nats
	if !opts.NoEvents {
		eventPub, err = events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
		}
	}

	// jwk
	jwks := jwk.NewJwks()
	err = jwks.AddKey(jwtKey.PublicKey, "key1")
	if err != nil {
		return nil, fmt.Errorf("failed to add key to JWKs: %w", err)
	}

	// store
	store, err := sqlite3.NewAuthStore(opts.SaveDir, opts.InMemory)
	if err != nil {
		return nil, err
	}

	wAuth, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "GoChat",                     // Display Name for your site
		RPID:          "localhost",                  // Generally the domain name for your site
		RPOrigins:     []string{"http://localhost"}, // The origin URL for WebAuthn requests
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &AuthService{
		store:    store,
		db:       store.DB(),
		jwtKey:   jwtKey,
		jwks:     jwks,
		wsTokens: cache.New(5*time.Minute, 10*time.Minute),
		eventPub: eventPub,
		wAuth:    wAuth,
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

func (s *AuthService) RegisterUser(ctx context.Context, username, password, name string) (*model.User, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err = credentialsRule(username, password); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrConstraint, err)
	}
	passwordHash, err := newHash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u, err := s.store.CreateUser(ctx, tx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %s: %w", username, err)
	}

	if err = s.store.SetPasswordHash(ctx, tx, u.Id, passwordHash); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if s.eventPub != nil {
		if err = s.eventPub.Publish(events.UserRegistered{
			UserId:    u.Id,
			Username:  name,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			log.Err(err).Msg("failed to publish user registered event")
		}
	}

	return u, nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, uid int32, password string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	passwordHash, err := newHash(password)
	if err != nil {
		return fmt.Errorf("failed to create hash password: %w", err)
	}
	if err = s.store.SetPasswordHash(ctx, tx, uid, passwordHash); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	// TODO: publish event

	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*model.Token, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	pw, err := s.store.GetPasswordByUsername(ctx, tx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get password hash: %w", err)
	}

	if match := pw.ValidatePassword(password); !match {
		return nil, errs.ErrAuthenticationFailure
	}

	// authentication success
	expiresIn := time.Hour * 24 * 7
	token, err := newToken(pw.UserId, s.jwtKey, expiresIn) // 7 days
	if err != nil {
		return nil, err
	}

	// TODO: publish event

	return token, nil
}

func (s *AuthService) WebauthnRegisterStart(ctx context.Context, userId int32) (*protocol.CredentialCreation, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := s.store.LoadWebAuthnUser(ctx, tx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to load webauthn user: %w", err)
	}

	creation, session, err := s.wAuth.BeginMediatedRegistration(u, protocol.MediationDefault,
		webauthn.WithExclusions(webauthn.Credentials(u.WebAuthnCredentials()).CredentialDescriptors()),
		webauthn.WithExtensions(map[string]any{"creProps": true}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	if err = s.store.SaveSessionData(ctx, tx, session); err != nil {
		return nil, fmt.Errorf("cannot save session data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return creation, nil
}

func (s *AuthService) WebauthnRegisterFinish(ctx context.Context, userId int32, pcc *protocol.ParsedCredentialCreationData) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := s.store.LoadWebAuthnUser(ctx, tx, userId)
	if err != nil {
		return fmt.Errorf("failed to load webauthn user: %w", err)
	}

	challenge := pcc.Response.CollectedClientData.Challenge

	session, err := s.store.GetSessionData(ctx, tx, challenge)
	if err != nil {
		return fmt.Errorf("failed to get session data: %w", err)
	}
	credential, err := s.wAuth.CreateCredential(u, *session, pcc)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}
	if err = s.store.SaveWebAuthnCredential(ctx, tx, userId, credential); err != nil {
		return fmt.Errorf("cannot save webauthn credential: %w", err)
	}

	if err = s.store.DeleteSessionData(ctx, tx, challenge); err != nil {
		return fmt.Errorf("cannot delete session data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// func (s *AuthService) GetUserByAccessToken(ctx context.Context, accessToken string) (*model.User, error) {
// 	tx, err := s.db.BeginTx(ctx, nil)
// 	uid, err := s.ValidateAccessToken(ctx, accessToken)
// 	if err != nil {
// 		return nil, err
// 	}
// 	u, err := s.store.GetUserById(ctx, tx, uid)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return u, nil
// }

// TODO: implement refresh token
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

func newToken(userId int32, jwtKey *rsa.PrivateKey, expiresIn time.Duration) (*model.Token, error) {
	accessToken, err := newClaims(userId, jwtKey, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}
	return &model.Token{
		UserId:       userId,
		AccessToken:  accessToken,
		RefreshToken: "", // TODO
		ExpiresIn:    expiresIn,
	}, nil
}

func newClaims(userId int32, jwtKey *rsa.PrivateKey, expiresIn time.Duration) (accessToken string, err error) {
	issudAt := time.Now()
	expiresAt := issudAt.Add(expiresIn)

	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": strconv.Itoa(int(userId)),
		"iat": jwt.NewNumericDate(issudAt),
		"exp": jwt.NewNumericDate(expiresAt),
		"aud": "api",
	})
	claims.Header["kid"] = "key1"

	accessToken, err = claims.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	log.Info().Msgf("Generated access token for user %d", userId)

	return accessToken, nil
}

func newHash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
