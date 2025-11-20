package presence

import (
	"net"
	"net/http"
)

func NewHttpServer(opts *HttpOptions) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	handler, err := NewHttpHandler(opts)

	if err != nil {
		return nil, err
	}
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return &srv, nil
}
