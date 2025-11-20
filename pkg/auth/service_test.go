package auth_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/elug3/gochat/pkg/auth"
	"github.com/elug3/gochat/pkg/auth/internal/errs"
	"github.com/elug3/gochat/pkg/auth/internal/model"
)

func newTestService(t *testing.T) (*auth.AuthService, error) {
	t.Helper()

	return auth.NewAuthService(&auth.Options{
		UseTmpKey: true,
		InMemory:  true,
		NoEvents:  true,
	})
}

func genUser(t *testing.T, service *auth.AuthService) (*model.User, error) {
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

// func TestWsToken(t *testing.T) {
// 	service, err := newTestService(t)
// 	if err != nil {
// 		t.Fatalf("failed to create test service: %v", err)
// 	}

// 	uid, wsToken, err := genTestAccessToken(t.Context(), service)
// 	if err != nil {
// 		t.Fatalf("failed to generate test access token: %v", err)
// 	}

// 	uid2, err := service.UseWsToken(t.Context(), wsToken)
// 	if err != nil {
// 		t.Errorf("failed to use WebSocket token: %v", err)
// 	}

// 	if uid != uid2 {
// 		t.Errorf("expected user Id %d, got %d", uid, uid2)
// 	}

// 	// check token volatility
// 	if _, err := service.UseWsToken(t.Context(), wsToken); err != nil {
// 		// TODO check error type
// 	} else {
// 		t.Errorf("token has not been revoked")
// 	}
// }

// func genTestAccessToken(ctx context.Context, s *auth.AuthService) (int32, string, error) {
// 	username := genRandomString()
// 	password := genRandomString()

// 	s.RegisterUser(ctx, username, password, "test")

// 	token, err := s.Login(ctx, username, password)
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	wsToken, err := s.CreateWsToken(ctx, token.AccessToken)
// 	if err != nil {
// 		return 0, "", err
// 	}

// 	return token.UserId, wsToken, nil
// }

// func genRandomBase32() string {
// 	b := make([]byte, 16)
// 	if _, err := rand.Read(b); err != nil {
// 		return ""
// 	}
// 	return base32.StdEncoding.EncodeToString(b)
// }

func genRandomUsername() string {
	num := rand.IntN(1000000)
	return fmt.Sprintf("user_%d", num)
}
