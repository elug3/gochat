package contacts

import (
	"context"
	"net"
	"net/http"

	"github.com/elug3/gochat/pkg/contacts/internal/store/s3"
	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

func NewHttpServer(opts *Options) (*http.Server, *ContactsService, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)

	contactsStore, isNewStore, err := sqlite3.NewContactsStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return nil, nil, err
	}
	var pub *events.Publisher
	if !opts.NoEvent {
		pub, err = events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return nil, nil, err
		}
	}
	var iconStore *s3.IconStore
	if !opts.NoIcons {
		iconStore, err = s3.NewIconStore(opts.S3Endpoint, opts.S3AccessKey, opts.S3SecretKey, opts.S3Region)
		if err != nil {
			return nil, nil, err
		}
	}
	s, err := NewContactsService(ServiceDeps{
		Store:       contactsStore,
		Publisher:   pub,
		IconStore:   iconStore,
		IconBaseURL: opts.IconBaseURL,
		IsNewStore:  isNewStore,
	})
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
