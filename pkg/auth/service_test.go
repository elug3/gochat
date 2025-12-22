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

	deps, err := auth.NewServiceDeps(&auth.Options{
		UseTmpKey: true,
		InMemory:  true,
		NoEvents:  true,
	})
	if err != nil {
		return nil, err
	}
	service, err := auth.NewAuthService(deps)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		service.Close()
	})

	return service, nil
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

func TestRegisterUserTwiceFails(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	username := genRandomUsername()
	password := "password"

	if _, err := service.RegisterUser(t.Context(), username, password, "test"); err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	if _, err := service.RegisterUser(t.Context(), username, password, "test"); errors.Is(err, errs.ErrConstraint) == false {
		t.Fatalf("expected ErrConstraint when registering duplicate user, got: %v", err)
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

func TestLoginAndValidateAccessToken(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	u, err := genUser(t, service)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, err := service.Login(t.Context(), u.Username, "password")
	if err != nil {
		t.Fatalf("failed to login with valid credentials: %v", err)
	}
	if token.UserId != u.Id {
		t.Fatalf("expected token user id %d, got %d", u.Id, token.UserId)
	}

	uid, err := service.ValidateAccessToken(t.Context(), token.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if uid != u.Id {
		t.Fatalf("expected validated user id %d, got %d", u.Id, uid)
	}

	if _, err := service.Login(t.Context(), u.Username, "wrong-password"); errors.Is(err, errs.ErrAuthenticationFailure) == false {
		t.Fatalf("expected authentication failure for wrong password, got: %v", err)
	}

	if _, err := service.ValidateAccessToken(t.Context(), token.AccessToken+"tampered"); err == nil {
		t.Fatalf("expected tampered token to be rejected")
	}
}

func TestWebsocketTokenLifecycle(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	u, err := genUser(t, service)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	token, err := service.Login(t.Context(), u.Username, "password")
	if err != nil {
		t.Fatalf("failed to login with valid credentials: %v", err)
	}

	wsToken, err := service.CreateWsToken(t.Context(), token.AccessToken)
	if err != nil {
		t.Fatalf("failed to create websocket token: %v", err)
	}
	if wsToken == "" {
		t.Fatalf("expected websocket token value")
	}

	userId, err := service.UseWsToken(t.Context(), wsToken)
	if err != nil {
		t.Fatalf("failed to use websocket token: %v", err)
	}
	if userId != u.Id {
		t.Fatalf("expected user id %d from websocket token, got %d", u.Id, userId)
	}

	if _, err := service.UseWsToken(t.Context(), wsToken); errors.Is(err, errs.ErrAuthenticationFailure) == false {
		t.Fatalf("expected websocket token to be single-use, got: %v", err)
	}
}

func TestUpdatePassword(t *testing.T) {
	service, err := newTestService(t)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}

	username := genRandomUsername()
	u, err := service.RegisterUser(t.Context(), username, "old-password", "test")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	if err := service.UpdatePassword(t.Context(), u.Id, "new-password"); err != nil {
		t.Fatalf("failed to update password: %v", err)
	}

	if _, err := service.Login(t.Context(), username, "old-password"); errors.Is(err, errs.ErrAuthenticationFailure) == false {
		t.Fatalf("expected login with old password to fail, got: %v", err)
	}

	if _, err := service.Login(t.Context(), username, "new-password"); err != nil {
		t.Fatalf("expected login with new password to succeed, got: %v", err)
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
