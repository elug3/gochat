package auth_test

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/elug3/gochat/pkg/auth"
	"github.com/elug3/gochat/pkg/auth/domain"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
)

var testOptions = auth.Options{
	UseTemporaryJWTKey: true,
	InMemory:           true,
	NoEvents:           true,
	RPDisplayName:      "test",
	RPID:               "",
	RPOrigins:          []string{"*"},
	LogLevel:           "none",
}

func newTestService(t *testing.T) (*auth.AuthService, error) {
	t.Helper()

	deps, err := auth.NewServerDeps(testOptions)
	if err != nil {
		return nil, fmt.Errorf("new server deps: %w", err)
	}
	service, err := auth.NewService(
		deps.Logger,
		deps.AuthStore,
		deps.JwtPrivKey,
		deps.Jwks,
		deps.WebAuthn,
		deps.Publisher,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func genUser(t *testing.T, service *auth.AuthService) (*domain.User, error) {
	t.Helper()

	username := genRandomUsername()
	password := "password"

	u, err := service.RegisterUser(t.Context(), username, password, "test")
	if err != nil {
		return nil, err
	}

	return u, nil
}

func TestCreateUser(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	testCase := map[string]struct {
		username string
		password string
		wantErr  error
	}{
		"create user": {
			username: "test",
			password: "password",
		},
		"empty password": {
			username: "test",
			password: "",
			wantErr:  errs.ErrConstraint,
		},
	}

	for name, tc := range testCase {
		t.Run(name, func(t *testing.T) {
			_, err := service.RegisterUser(t.Context(), tc.username, tc.password, "test")
			if errors.Is(err, tc.wantErr) == false {
				t.Errorf("expected error: %v, got: %v", tc.wantErr, err)
			}
		})

	}
}

func TestPassword(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	testCases := map[string]struct {
		setPassword   string
		inputPassword string
		wantErr       error
	}{
		"normal password": {setPassword: "password", inputPassword: "password"},
		"empty password":  {setPassword: "", inputPassword: "", wantErr: errs.ErrConstraint},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := service.RegisterUser(t.Context(), genRandomUsername(), tc.setPassword, "test")
			if errors.Is(err, tc.wantErr) == false {
				t.Fatalf("failed to create user: %v", err)
			}
		})
	}
}

func genRandomUsername() string {
	num := rand.IntN(1000000)
	return fmt.Sprintf("user_%d", num)
}

func hashSessionIDForTest(sessionID string) (string, error) {
	hasher := sha512.New512_256()
	if _, err := hasher.Write([]byte(sessionID)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func timesClose(a, b time.Time) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Second
}
