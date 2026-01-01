package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/elug3/gochat/pkg/auth/internal/jwk"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3/authstore"
	"github.com/elug3/gochat/pkg/auth/internal/store/sqlite3/outboxstore"
	"github.com/elug3/gochat/shared/events"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog"
)

type Server struct {
	logger  *zerolog.Logger
	httpSrv *http.Server
	wg      sync.WaitGroup
	closers []func() error
}

func NewServer(opts Options) (*Server, error) {
	deps, err := NewServerDeps(opts)
	if err != nil {
		return nil, fmt.Errorf("new server deps: %w", err)
	}

	return &Server{
		logger:  deps.Logger,
		httpSrv: deps.HttpSrv,
	}, nil
}

func (srv *Server) Run(ctx context.Context) error {
	srv.logger.Info().Msg("server starting...")
	errCh := make(chan error)

	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()
		if err := srv.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			srv.logger.Err(err).Send()
		}
	}

	shutdowwnCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.httpSrv.Shutdown(shutdowwnCtx); err != nil {
		return err
	}

	srv.wg.Wait()
	return nil
}

type initStep func(deps *ServerDeps) error

type ServerDeps struct {
	Logger      *zerolog.Logger
	JwtPrivKey  *rsa.PrivateKey
	HttpSrv     *http.Server
	Jwks        *jwk.Jwks
	WebAuthn    *webauthn.WebAuthn
	Publisher   *events.Publisher
	DB          *sql.DB
	AuthStore   *authstore.Store
	OutboxStore *outboxstore.Store
	Auth        *AuthService
	Handler     *AuthHandler
	closers     []func() error
}

func NewServerDeps(opts Options) (*ServerDeps, error) {
	steps := []initStep{
		initLogger(opts),
		initJWTKey(opts),
		initJwks(opts),
		initWebauthn(opts),
		initEvent(opts),
		initOpenDB(opts),
		initAuthStore(opts),
		initOutboxStore(opts),
		initAuthService(opts),
		initHandler(opts),
		initHttpServer(opts),
	}
	deps := ServerDeps{}
	for _, step := range steps {
		if err := step(&deps); err != nil {
			deps.CleanupDeps()
			return nil, fmt.Errorf("init: %w", err)
		}
	}
	return &deps, nil
}

func initLogger(opts Options) initStep {
	return func(deps *ServerDeps) error {
		var (
			logLevel zerolog.Level
			logger   zerolog.Logger
		)
		switch opts.LogLevel {
		case "debug":
			logLevel = zerolog.DebugLevel
		case "warn":
			logLevel = zerolog.WarnLevel
		case "error":
			logLevel = zerolog.ErrorLevel
		case "info":
			logLevel = zerolog.InfoLevel
		case "none":
			logLevel = zerolog.Disabled
		default:
			return fmt.Errorf("invalid log level: %s", opts.LogLevel)
		}

		switch opts.LogFormat {
		case "text":
			logger = zerolog.New(zerolog.NewConsoleWriter())
		case "json":
			logger = zerolog.New(os.Stdout).Level(logLevel)
		default:
			return fmt.Errorf("invalid log format: %s", opts.LogFormat)
		}

		deps.Logger = &logger
		return nil
	}
}

func initAuthService(opts Options) initStep {
	return func(deps *ServerDeps) error {
		auth, err := NewService(
			deps.Logger,
			deps.AuthStore,
			deps.JwtPrivKey,
			deps.Jwks,
			deps.WebAuthn,
			deps.Publisher,
		)
		if err != nil {
			return err
		}

		deps.Auth = auth
		return nil
	}
}

func initHandler(opts Options) initStep {
	return func(deps *ServerDeps) error {
		deps.Handler = newAuthHandler(deps.Auth)
		return nil
	}
}

func initHttpServer(opts Options) initStep {
	return func(deps *ServerDeps) error {
		addr := net.JoinHostPort(opts.Host, opts.Port)
		deps.HttpSrv = &http.Server{
			Addr:    addr,
			Handler: deps.Handler,
		}
		return nil
	}
}

func initJWTKey(opts Options) initStep {
	return func(deps *ServerDeps) error {
		if !opts.UseTemporaryJWTKey && opts.JWTKeyPath == "" {
			return fmt.Errorf("%w: jwt private key path is required when temporary key is disabled", ErrInvalidConfig)
		}
		var (
			privKey *rsa.PrivateKey
			err     error
		)
		if opts.UseTemporaryJWTKey {
			if privKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
				return fmt.Errorf("rsa.GenerateKey: %w", err)
			}
		} else {
			keyData, err := os.ReadFile(opts.JWTKeyPath)
			if err != nil {
				return fmt.Errorf("read key %q: %w", opts.JWTKeyPath, err)
			}

			der, _ := pem.Decode(keyData)
			key, err := x509.ParsePKCS8PrivateKey(der.Bytes)
			if err != nil {
				return fmt.Errorf("x509.ParsePKCS8PrivateKey: %w", err)
			}
			var ok bool
			if privKey, ok = key.(*rsa.PrivateKey); !ok {
				return errors.New("not an RSA private key")
			}
		}
		deps.JwtPrivKey = privKey
		return nil
	}
}

func initEvent(opts Options) initStep {
	return func(deps *ServerDeps) error {
		if opts.NoEvents {
			return nil
		}

		publisher, err := events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return err
		}

		deps.Publisher = publisher
		deps.closers = append(deps.closers, func() error {
			publisher.Close()
			return nil
		})
		return nil
	}

}

func initWebauthn(opts Options) initStep {
	return func(deps *ServerDeps) error {
		webAuthn, err := webauthn.New(&webauthn.Config{
			RPDisplayName: opts.RPDisplayName,
			RPID:          opts.RPID,
			RPOrigins:     opts.RPOrigins,
		})
		if err != nil {
			return err
		}
		deps.WebAuthn = webAuthn
		return nil
	}
}

func initOpenDB(opts Options) initStep {
	return func(deps *ServerDeps) error {
		var (
			db  *sql.DB
			err error
		)
		if opts.InMemory {
			if db, err = sql.Open("sqlite3", ":memory:"); err != nil {
				return err
			}
		} else {
			path := opts.SaveDir + "/auth.db"
			if db, err = sql.Open("sqlite3", path); err != nil {
				return err
			}
		}
		deps.DB = db
		return nil
	}
}

func initAuthStore(opts Options) initStep {
	return func(deps *ServerDeps) error {
		store, err := authstore.NewAuthStore(context.TODO(), deps.DB)
		if err != nil {
			return err
		}
		deps.AuthStore = store
		return nil
	}
}

func initOutboxStore(opts Options) initStep {
	return func(deps *ServerDeps) error {
		store, err := outboxstore.NewStore(context.TODO(), deps.DB)
		if err != nil {
			return nil
		}
		deps.OutboxStore = store
		return nil
	}
}

func initJwks(opts Options) initStep {
	return func(deps *ServerDeps) error {
		deps.Jwks = jwk.NewJwks()
		return nil
	}
}

func (deps *ServerDeps) CleanupDeps() {
	for i := len(deps.closers) - 1; i >= 0; i-- {
		_ = deps.closers[i]()
	}
	deps.closers = nil
}

func (srv *Server) Close() error {
	if srv == nil {
		return nil
	}
	var closeErr error
	for i := len(srv.closers) - 1; i >= 0; i-- {
		if err := srv.closers[i](); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (srv *Server) Shutdown(ctx context.Context) {
}
