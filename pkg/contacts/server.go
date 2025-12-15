package contacts

import (
	"context"
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

func NewHttpServer(opts *Options) (*http.Server, *ContactsService, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewContactsService(opts)
	if err != nil {
		return nil, nil, err
	}
	h := newContactsHandler(s)

	server := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}
	return server, s, nil
}

// Shutdown gracefully shuts down the HTTP server
func ShutdownServer(ctx context.Context, srv *http.Server) error {
	log.Info().Msg("shutting down HTTP server")
	return srv.Shutdown(ctx)
}

// HealthCheck checks the health of the service and its dependencies
func (s *ContactsService) HealthCheck(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})
	health["status"] = "ok"

	// Check database connectivity
	tx, err := s.store.Begin()
	if err != nil {
		health["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		tx.Rollback()
		health["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check store type for additional info
	health["store_type"] = "sqlite3" // Currently only SQLite is supported

	health["event_publisher"] = s.pub != nil
	health["icon_store"] = s.iconStore != nil

	return health
}
