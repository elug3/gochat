package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/coder/websocket"
)

// UserIdKey is the context key for user Id
type UserIdKey struct{}

type WsServer struct {
	addr string
	mux  http.ServeMux
	auth *JwkAuthenticator
	hub  *Hub
}

func NewWsServer(hub *Hub, opts *Options) (*WsServer, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	auth, err := NewAuthenticator(opts.AuthUrl)
	if err != nil {
		return nil, err
	}
	s := WsServer{
		addr: addr,
		auth: auth,
		hub:  hub,
	}

	s.mux.HandleFunc("/ws", s.AuthMiddleware(s.ServeWs))

	return &s, nil
}

func (s *WsServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := parseToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userId, err := s.auth.Authenticate(r.Context(), token)
		if err != nil || userId == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// You can set the user ID in the request context if needed
		ctx := context.WithValue(r.Context(), UserIdKey{}, userId)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

func (s *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *WsServer) ListenAndServe() error {
	return http.ListenAndServe(s.addr, &s.mux)
}

func (s *WsServer) ServeWs(w http.ResponseWriter, r *http.Request) {
	userId := getUserIdFromContext(r.Context())
	if userId == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	client := NewClient(userId, conn)
	s.hub.Register(client)
}

// parseToken extracts the token from the Authorization header
func parseToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	s := strings.SplitN(auth, " ", 2)
	if len(s) != 2 || s[0] != "Bearer" {
		return "", fmt.Errorf("invalid token")
	}
	return s[1], nil
}

// getUserIdFromContext extracts the user ID from the context
func getUserIdFromContext(ctx context.Context) int32 {
	if ctx == nil {
		return 0
	}
	userId, ok := ctx.Value(UserIdKey{}).(int32)
	if !ok {
		return 0
	}
	return userId
}
