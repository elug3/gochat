package webauthn

import (
	"net"
	"net/http"

	"github.com/rs/cors"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	mux := http.NewServeMux()
	addr := net.JoinHostPort(opts.Host, opts.Port)
	h, err := NewHandler(opts)
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/register/start", h.registerStart)
	mux.HandleFunc("/register/finish", h.registerFinish)

	mux.HandleFunc("/login/start", h.loginStart)
	mux.HandleFunc("/login/finish", h.loginFinish)

	mux.HandleFunc("/registered-credentials", h.registeredCredentials)

	c := cors.New(cors.Options{
		AllowedOrigins:   opts.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}, nil
}
