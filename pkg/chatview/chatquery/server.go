package chatquery

import (
	"net"
	"net/http"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewChatViewService(opts)
	if err != nil {
		return nil, err
	}
	srv := &http.Server{
		Addr:    addr,
		Handler: NewHandler(s),
	}
	return srv, nil
}
