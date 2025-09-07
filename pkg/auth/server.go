package auth

import (
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

type HttpServer struct {
	httpSrv *http.Server
}

func NewHttpServer(opts *Options) (*HttpServer, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewAuthService(opts)
	if err != nil {
		return nil, err
	}
	srv := HttpServer{
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: newAuthHandler(s),
		},
	}
	return &srv, nil
}

func (s *HttpServer) Run() error {
	log.Info().Msgf("server started on %s", s.httpSrv.Addr)
	return s.httpSrv.ListenAndServe()
}
