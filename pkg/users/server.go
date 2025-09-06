package users

import (
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

type HttpServer struct {
	httpSrv *http.Server
	users   *UserService
}

func NewHttpServer(opts *Options) (*HttpServer, error) {
	s, err := NewUserService()
	if err != nil {
		return nil, err
	}

	addr := net.JoinHostPort(opts.Host, opts.Port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: newUserHandler(s),
	}

	return &HttpServer{
		httpSrv: httpSrv,
		users:   s,
	}, nil
}

func (srv *HttpServer) ListenAndServe() error {
	return srv.httpSrv.ListenAndServe()
}

func (srv *HttpServer) Run() error {
	log.Info().Msgf("starting user service server on %s", srv.httpSrv.Addr)
	return srv.ListenAndServe()
}
