package presence

import (
	"net"
	"net/http"

	"github.com/elug3/gochat/pkg/presence/internal/store/redisstore"
)

func NewHttpServer(opts *HttpOptions) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	store, err := redisstore.NewPresenceStore(opts.RedisAddr)
	if err != nil {
		return nil, err
	}
	presences, err := NewPresenceService(store)
	if err != nil {
		return nil, err
	}
	handler, err := NewHttpHandler(presences)

	if err != nil {
		return nil, err
	}
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return &srv, nil
}
