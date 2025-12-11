package contacts

import (
	"net"
	"net/http"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewContactsService(opts)
	if err != nil {
		return nil, err
	}
	h := newContactsHandler(s)

	return &http.Server{
		Addr:    addr,
		Handler: h,
	}, nil
}
