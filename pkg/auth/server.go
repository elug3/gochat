package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/elug3/gochat/shared/events"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog"
)

type Server struct {
	privKey   *rsa.PrivateKey
	logger    *zerolog.Logger
	webAuthn  *webauthn.WebAuthn
	publisher *events.Publisher
	closers   []func() error
}

func NewServer(opts Options) (*Server, error) {
	steps := map[string]initStep{
		"logger":   initLogger(opts),
		"jwt-key":  initJWTKey(opts),
		"webauthn": initWebauthn(opts),
		"nats":     initEvent(opts),
	}
	deps := serverDeps{}
	for name, step := range steps {
		if err := step(&deps); err != nil {
			cleanupServerDeps(&deps)
			return nil, fmt.Errorf("init %s: %w", name, err)
		}
	}
	return &Server{
		privKey:   deps.jwtPrivKey,
		logger:    deps.logger,
		webAuthn:  deps.webAuthn,
		publisher: deps.publisher,
		closers:   deps.closers,
	}, nil
}

type initStep func(*serverDeps) error

type serverDeps struct {
	logger     *zerolog.Logger
	jwtPrivKey *rsa.PrivateKey
	webAuthn   *webauthn.WebAuthn
	closers    []func() error
	publisher  *events.Publisher
}

func initLogger(opts Options) initStep {
	return func(d *serverDeps) error {
		logger := zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger().
			Level(opts.LogLevel)
		d.logger = &logger
		return nil
	}
}

func initJWTKey(opts Options) initStep {
	return func(d *serverDeps) error {
		if !opts.UseTmpKey && opts.KeyPath == "" {
			return fmt.Errorf("%w: jwt private key path is required when temporary key is disabled", ErrInvalidConfig)
		}
		var (
			privKey *rsa.PrivateKey
			err     error
		)
		if opts.UseTmpKey {
			if privKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
				return fmt.Errorf("rsa.GenerateKey: %w", err)
			}
		} else {
			keyData, err := os.ReadFile(opts.KeyPath)
			if err != nil {
				return fmt.Errorf("read key %q: %w", opts.KeyPath, err)
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
		d.jwtPrivKey = privKey
		return nil
	}
}

func initEvent(opts Options) initStep {
	return func(d *serverDeps) error {
		if opts.NoEvents {
			return nil
		}

		publisher, err := events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return err
		}

		d.publisher = publisher
		d.closers = append(d.closers, func() error {
			publisher.Close()
			return nil
		})
		return nil
	}

}

func initWebauthn(opts Options) initStep {
	return func(d *serverDeps) error {
		webAuthn, err := webauthn.New(&webauthn.Config{
			RPDisplayName: opts.RPDisplayName,
			RPID:          opts.RPID,
			RPOrigins:     opts.RPOrigins,
		})
		if err != nil {
			return err
		}
		d.webAuthn = webAuthn
		return nil
	}
}

func cleanupServerDeps(d *serverDeps) {
	for i := len(d.closers) - 1; i >= 0; i-- {
		_ = d.closers[i]()
	}
	d.closers = nil
}

func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	var closeErr error
	for i := len(s.closers) - 1; i >= 0; i-- {
		if err := s.closers[i](); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}
