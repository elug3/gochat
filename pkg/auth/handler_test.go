package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elug3/gochat/pkg/auth/internal/model"
)

type meResponse struct {
	UserID int32 `json:"user_id"`
}

var testOptions = Options{
	UseTemporaryJWTKey: true,
	InMemory:           true,
	NoEvents:           true,
	LogLevel:           "none",
	RPDisplayName:      "test",
	RPID:               "localhost",
	RPOrigins:          []string{"http://localhost:8080"},
}

func newTestHTTPServer(t *testing.T) (*Server, error) {
	t.Helper()

	srv, err := NewServer(testOptions)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func registerViaHTTP(t *testing.T, client *http.Client, baseURL, username, password string) {
	t.Helper()

	body := fmt.Sprintf(`{"username":"%s","name":"test user","password":"%s"}`, username, password)
	resp, err := client.Post(baseURL+"/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected register status 200, got %d: %s", resp.StatusCode, string(b))
	}
}

func loginViaHTTP(t *testing.T, client *http.Client, baseURL, username, password string) *model.Session {
	t.Helper()
	body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/login", strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build login request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected login status 200, got %d: %s", resp.StatusCode, string(b))
	}

	var session model.Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if session.SessionId == "" {
		t.Fatalf("expected session id in login response")
	}

	return &session
}

// func TestAuthHandler_RegisterLoginAndMeFlow(t *testing.T) {
// 	server := newTestHTTPServer(t)
// 	client := server.Client()

// 	username := fmt.Sprintf("handler_user_%d", time.Now().UnixNano())
// 	password := "password123"

// 	registerViaHTTP(t, client, server.URL, username, password)
// 	loginResp := loginViaHTTP(t, client, server.URL, username, password)

// 	req, err := http.NewRequest(http.MethodGet, server.URL+"/me", http.NoBody)
// 	if err != nil {
// 		t.Fatalf("failed to build /me request: %v", err)
// 	}
// 	req.Header.Set("")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		t.Fatalf("/me request failed: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		b, _ := io.ReadAll(resp.Body)
// 		t.Fatalf("expected /me status 200, got %d: %s", resp.StatusCode, string(b))
// 	}

// 	var mr meResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
// 		t.Fatalf("failed to decode /me response: %v", err)
// 	}

// 	if mr.UserID != loginResp.UserId {
// 		t.Fatalf("expected /me user_id %d, got %d", loginResp.UserId, mr.UserID)
// 	}
// }

// func TestAuthHandler_MeRequiresValidToken(t *testing.T) {
// 	server := newTestHTTPServer(t)
// 	client := server.Client()

// 	t.Run("missing token", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, server.URL+"/me", http.NoBody)
// 		if err != nil {
// 			t.Fatalf("failed to build /me request: %v", err)
// 		}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			t.Fatalf("/me request failed: %v", err)
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode != http.StatusBadRequest {
// 			t.Fatalf("expected missing token to return 400, got %d", resp.StatusCode)
// 		}
// 	})

// 	t.Run("invalid token", func(t *testing.T) {
// 		req, err := http.NewRequest(http.MethodGet, server.URL+"/me", http.NoBody)
// 		if err != nil {
// 			t.Fatalf("failed to build /me request: %v", err)
// 		}
// 		req.Header.Set("Authorization", "Bearer not-a-token")

// 		resp, err := client.Do(req)
// 		if err != nil {
// 			t.Fatalf("/me request failed: %v", err)
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode != http.StatusUnauthorized {
// 			b, _ := io.ReadAll(resp.Body)
// 			t.Fatalf("expected invalid token to return 401, got %d: %s", resp.StatusCode, string(b))
// 		}
// 	})
// }
