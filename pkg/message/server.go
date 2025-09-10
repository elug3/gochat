package message

import (
	"net"
	"net/http"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewMessageService(opts)
	if err != nil {
		return nil, err
	}
	h := NewMessageHandler(s)
	server := &http.Server{
		Addr:    addr,
		Handler: h,
	}
	return server, nil
}
