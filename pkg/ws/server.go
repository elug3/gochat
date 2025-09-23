package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/coder/websocket"
	"github.com/elug3/gochat/api/httpclient"
	"github.com/rs/zerolog/log"
)

// UserIdKey is the context key for user Id
type UserIdKey struct{}

type WsServer struct {
	addr     string
	mux      http.ServeMux
	auth     *HttpAuthenticator
	Contacts *httpclient.ContactsClient
	hub      *Hub
}

func NewWsServer(hub *Hub, opts *Options) (*WsServer, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)

	auth, err := NewHttpAuthenticator(opts.AuthServerUrl)
	if err != nil {
		return nil, err
	}

	contactsClient := httpclient.NewContactsClient(
		httpclient.WithBaseUrl(opts.ContactsServerUrl),
	)

	s := WsServer{
		addr:     addr,
		auth:     auth,
		hub:      hub,
		Contacts: contactsClient,
	}

	s.mux.HandleFunc("/ws", s.AuthMiddleware(s.ServeWs))

	return &s, nil
}

func (s *WsServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsToken, err := parseTokenQuery(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		userId, err := s.auth.ValidateWsToken(r.Context(), wsToken)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIdKey{}, userId)
		next(w, r.WithContext(ctx))
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

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to accept websocket connection")
		return
	}

	s.hub.registerCh <- &RegisterMsg{UserId: userId, Conn: conn}
	groups, err := s.Contacts.Groups.ListByUser(r.Context(), userId)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get groups for user %d", userId)
	}
	for _, group := range groups {
		s.hub.subscribeCh <- &SubscribeMsg{ChatId: group.Id, UserId: userId}
	}
}

func getUserIdFromContext(ctx context.Context) int32 {
	if ctx == nil {
		return 0
	}
	if userId, ok := ctx.Value(UserIdKey{}).(int32); ok {
		return userId
	}
	return 0
}

func parseTokenQuery(r *http.Request) (string, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return "", fmt.Errorf("missing token in query")
	}
	return token, nil
}
