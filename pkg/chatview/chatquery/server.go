package chatquery

import (
	"net"
	"net/http"

	redisstore "github.com/elug3/gochat/pkg/chatview/chatquery/internal/store/redis-store"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	store, err := redisstore.NewChatViewStore(opts.DatabaseURL)
	if err != nil {
		return nil, err
	}
	s, err := NewChatViewService(ServiceDeps{Store: store})
	if err != nil {
		return nil, err
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: NewHandler(s),
	}
	return srv, nil
}
