package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/elug3/gochat/pkg/auth/domain"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/jwk"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3/authstore"
	"github.com/elug3/gochat/shared/events"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	logger       *zerolog.Logger
	authStore    *authstore.Store
	jwtKey       *rsa.PrivateKey
	jwks         *jwk.Jwks
	webAuthn     *webauthn.WebAuthn
	wsTokens     cache.Cache
	publisher    *events.Publisher
	outboxWorker *worker
}

type Client struct {
	IP        net.IP
	UserAgent string
}

var (
	AccessTokenExpiry  = 8 * time.Hour      // 8 hours
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

func NewService(
	logger *zerolog.Logger,
	authStore *authstore.Store,
	jwtKey *rsa.PrivateKey,
	jwks *jwk.Jwks,
	webAuthn *webauthn.WebAuthn,
	publisher *events.Publisher,
) (*AuthService, error) {
	if logger == nil {
		return nil, fmt.Errorf("missing logger")
	}
	if authStore == nil {
		return nil, fmt.Errorf("missing auth store")
	}
	if jwtKey == nil {
		return nil, fmt.Errorf("missing jwt key")
	}
	if jwks == nil {
		return nil, fmt.Errorf("missing jwks")
	}
	if webAuthn == nil {
		return nil, fmt.Errorf("missing webAuthn")
	}

	return &AuthService{
		logger:    logger,
		authStore: authStore,
		jwtKey:    jwtKey,
		jwks:      jwks,
		webAuthn:  webAuthn,
		publisher: publisher,
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

func (s *AuthService) RegisterUser(ctx context.Context, username, password, name string) (*domain.User, error) {
	tx, err := s.authStore.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err = credentialsRule(username, password); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrConstraint, err)
	}
	passwordHash, err := generatePasswordHash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.authStore.CreateUser(ctx, tx, username)
	if err != nil {
		return nil, fmt.Errorf("create user %q: %w", username, err)
	}

	if err = s.authStore.SetPasswordHash(ctx, tx, u.Id, passwordHash); err != nil {
		return nil, fmt.Errorf("set password hash user %d: %w", u.Id, err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	registeredEvent := events.UserRegistered{
		UserId:    u.Id,
		Username:  u.Username,
		Name:      name,
		Timestamp: time.Now().Unix(),
	}
	if s.outboxWorker != nil {
		if err = s.outboxWorker.Publish(ctx, registeredEvent); err != nil {
			return nil, fmt.Errorf("publish with outbox worker: %w", err)
		}
	} else if s.publisher != nil {
		if err = s.publisher.Publish(registeredEvent); err != nil {
			log.Err(err).Msg("failed to publish user registered event")
		}
	}

	return u, nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, uid int32, password string) error {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	passwordHash, err := generatePasswordHash(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err = s.authStore.SetPasswordHash(ctx, tx, uid, passwordHash); err != nil {
		return fmt.Errorf("set password hash user %d: %w", uid, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// TODO: publish event

	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string, client *Client) (*domain.Session, error) {
	if client == nil {
		client = &Client{
			IP:        net.IPv4zero,
			UserAgent: "Unknown",
		}
	}
	sessionDuration := 7 * 24 * time.Hour // TODO: make configurable
	expiresAt := time.Now().Add(sessionDuration)

	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	pw, err := s.authStore.GetPasswordByUsername(ctx, tx, username)
	if err != nil {
		return nil, fmt.Errorf("get password hash %q: %w", username, err)
	}

	if match := pw.ValidatePassword(password); !match {
		return nil, errs.ErrAuthenticationFailure
	}

	sessionId, err := generateSessionId()
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}
	sessionHash, err := hashSessionId(sessionId)
	if err != nil {
		return nil, fmt.Errorf("hash session id: %w", err)
	}

	createdAt := time.Now()
	if err = s.authStore.SaveSession(ctx, tx, pw.UserId, sessionHash, client.IP, client.UserAgent, createdAt, expiresAt); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	if s.publisher != nil {
		if err = s.publisher.Publish(events.UserLoggedIn{
			UserId:    pw.UserId,
			Username:  username,
			IP:        client.IP,
			UserAgent: client.UserAgent,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			log.Err(err).Msg("failed to publish user logged in event")
		}
	}

	return &domain.Session{
		SessionId: sessionId,
		UserId:    pw.UserId,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		IP:        client.IP,
		UserAgent: client.UserAgent,
	}, nil
}

func (s *AuthService) Exchange(ctx context.Context, sessionId string, requestedAudiences, requestedScopes []string, requestedTTL int64) (*domain.Token, error) {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	session, err := s.validateSession(ctx, tx, sessionId)
	if err != nil {
		return nil, fmt.Errorf("validate session %q: %w", sessionId, err)
	}

	token, err := newToken(session.UserId, s.jwtKey, requestedAudiences, requestedTTL)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return token, nil
}

func (s *AuthService) validateSession(ctx context.Context, tx *sql.Tx, sessionId string) (*domain.Session, error) {
	hash, err := hashSessionId(sessionId)
	if err != nil {
		return nil, fmt.Errorf("hash session id: %w", err)
	}
	session, err := s.authStore.GetSessionByHash(ctx, tx, hash)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errs.ErrSessionExpired
	}
	return session, nil
}

func (s *AuthService) WebAuthnRegisterBegin(ctx context.Context, userId int32) (*protocol.CredentialCreation, error) {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	u, err := s.authStore.GetWebAuthnUser(ctx, tx, userId)
	if err != nil {
		return nil, fmt.Errorf("load webauthn user %d: %w", userId, err)
	}

	creation, session, err := s.webAuthn.BeginMediatedRegistration(u, protocol.MediationOptional,
		webauthn.WithExclusions(webauthn.Credentials(u.WebAuthnCredentials()).CredentialDescriptors()),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		webauthn.WithExtensions(map[string]any{"creProps": true}),
	)
	if err != nil {
		return nil, fmt.Errorf("begin registration: %w", err)
	}

	if err = s.authStore.SaveWebauthnSessionData(ctx, tx, session); err != nil {
		return nil, fmt.Errorf("save session data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return creation, nil
}

func (s *AuthService) WebAuthnRegisterFinish(ctx context.Context, userId int32, pcc *protocol.ParsedCredentialCreationData) error {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	credentialName := "my_passkey" // TODO: get from request

	u, err := s.authStore.GetWebAuthnUser(ctx, tx, userId)
	if err != nil {
		return fmt.Errorf("load webauthn user %d: %w", userId, err)
	}

	challenge := pcc.Response.CollectedClientData.Challenge

	session, err := s.authStore.GetWebauthnSessionData(ctx, tx, challenge)
	if err != nil {
		return fmt.Errorf("get session data: %w", err)
	}
	credential, err := s.webAuthn.CreateCredential(u, *session, pcc)
	if err != nil {
		return fmt.Errorf("create credential: %w", err)
	}

	if err = s.authStore.SaveWebAuthnCredential(ctx, tx, userId, credentialName, credential); err != nil {
		return fmt.Errorf("save webauthn credential: %w", err)
	}

	if err = s.authStore.DeleteWebauthnSessionData(ctx, tx, challenge); err != nil {
		return fmt.Errorf("delete session data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	log.Info().Msgf("Registered new passkey for user %d", userId)

	return nil
}

func (s *AuthService) WebAuthnLoginBegin(ctx context.Context) (*protocol.CredentialAssertion, error) {
	assertion, session, err := s.webAuthn.BeginDiscoverableMediatedLogin(protocol.MediationOptional)
	if err != nil {
		return nil, fmt.Errorf("begin login: %w", err)
	}
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err = s.authStore.SaveWebauthnSessionData(ctx, tx, session); err != nil {
		return nil, fmt.Errorf("save session data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return assertion, nil
}

func (s *AuthService) WebAuthnLoginFinish(ctx context.Context, parsedResponse *protocol.ParsedCredentialAssertionData, client *Client) (*domain.Session, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // TODO: make configurable
	var (
		u   *domain.WebAuthnUser
		err error
	)
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	challenge := parsedResponse.Response.CollectedClientData.Challenge

	webauthnSession, err := s.authStore.GetWebauthnSessionData(ctx, tx, challenge)
	if err != nil {
		return nil, fmt.Errorf("get session data: %w", err)
	}

	credential, err := s.webAuthn.ValidateDiscoverableLogin(func(rawID, userHandle []byte) (user webauthn.User, err error) {
		uid, err := domain.UserIdFromUserHandler(userHandle)
		if err != nil {
			return nil, fmt.Errorf("invalid user handle: %w", err)
		}
		if u, err = s.authStore.GetWebAuthnUser(ctx, tx, uid); err != nil {
			return nil, fmt.Errorf("load webauthn user %d: %w", uid, err)
		}
		return u, nil
	}, *webauthnSession, parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("validate login: %w", err)
	}

	if err = s.authStore.UpdateWebAuthnCredentialAfterLogin(ctx, tx, u.Id, credential); err != nil {
		return nil, fmt.Errorf("update webauthn credential: %w", err)
	}

	if err = s.authStore.DeleteWebauthnSessionData(ctx, tx, challenge); err != nil {
		return nil, fmt.Errorf("delete session data: %w", err)
	}

	sessionId, err := generateSessionId()
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}
	sessionHsah, err := hashSessionId(sessionId)
	if err != nil {
		return nil, fmt.Errorf("hash session id: %w", err)
	}

	createdAt := time.Now()
	if err = s.authStore.SaveSession(ctx, tx, u.Id, sessionHsah, client.IP, client.UserAgent, createdAt, expiresAt); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &domain.Session{
		SessionId: sessionId,
		UserId:    u.Id,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		IP:        client.IP,
		UserAgent: client.UserAgent,
	}, nil
}

func (s *AuthService) GetWebAuthnUserPasskeys(ctx context.Context, userId int32) ([]domain.Passkey, error) {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	passkeys, err := s.authStore.GetPasskeysByUserId(ctx, tx, userId)
	if err != nil {
		return nil, fmt.Errorf("get passkeys: %w", err)
	}

	return passkeys, nil
}

func (s *AuthService) UpdateWebAuthnPasskey(ctx context.Context, userId, passkeyId int32, newName string) (*domain.Passkey, error) {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	passkey, err := s.authStore.UpdatePasskey(ctx, tx, passkeyId, newName)
	if err != nil {
		return nil, fmt.Errorf("update passkey %d: %w", passkeyId, err)
	}
	if passkey.UserId != userId {
		return nil, fmt.Errorf("passkey %d does not belong to user", passkeyId)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	log.Info().Msgf("Updated passkey %d for user %d", passkeyId, userId)
	return passkey, nil
}

func (s *AuthService) DeleteWebAuthnPasskey(ctx context.Context, userId, passkeyId int32) error {
	tx, err := s.authStore.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	passkey, err := s.authStore.DeletePasskeyById(ctx, tx, passkeyId)
	if err != nil {
		return fmt.Errorf("delete passkey %d: %w", passkeyId, err)
	}
	if passkey.UserId != userId {
		return fmt.Errorf("passkey %d does not belong to user", passkeyId)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	log.Info().Msgf("Deleted passkey %d for user %d", passkeyId, userId)
	return nil
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
		return 0, fmt.Errorf("parse token: %w", err)
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

func (s *AuthService) CreateWsToken(ctx context.Context, accessToken string) (string, error) {
	userId, err := s.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		return "", err
	}
	return s.createWsToken(ctx, userId)
}

func (s *AuthService) createWsToken(ctx context.Context, userId int32) (string, error) {
	token := rand.Text()
	s.wsTokens.Set(token, userId, time.Second*300) // 5 minute

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
	return s.authStore.Close()
}

func newToken(userId int32, jwtKey *rsa.PrivateKey, audiences []string, TTL int64) (*domain.Token, error) {
	accessToken, err := newClaims(userId, jwtKey, audiences, TTL)
	if err != nil {
		return nil, err
	}
	return &domain.Token{
		UserId:      userId,
		AccessToken: accessToken,
	}, nil
}

func newClaims(userId int32, jwtKey *rsa.PrivateKey, audiences []string, TTL int64) (accessToken string, err error) {
	issudAt := time.Now().Unix()
	expiresAt := issudAt + TTL
	userIdStr := strconv.FormatInt(int64(userId), 10)

	claims := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "auth",
		"sub": userIdStr,
		"aud": audiences,
		"iat": issudAt,
		"exp": expiresAt,
		"nbf": issudAt,
	})
	// TODO: add iti, client_id, scp field later
	claims.Header["kid"] = "key1"

	accessToken, err = claims.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func generatePasswordHash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func generateTokenString(prefix string) (string, error) {
	length := 16
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, base32.StdEncoding.EncodeToString(b)), nil
}

func hashSessionId(s string) (string, error) {
	hasher := sha512.New512_256()
	_, err := hasher.Write([]byte(s))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func generateSessionId() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	sessionId := fmt.Sprintf("sess_%s", base32.StdEncoding.EncodeToString(b))

	return sessionId, nil
}
