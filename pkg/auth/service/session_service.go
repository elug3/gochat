package service

import "github.com/elug3/gochat/pkg/auth/ports"

// SessionService manages auth session lifecycle operations.
type SessionService struct {
	sessions ports.SessionRepository
	signer   ports.TokenSigner
}

func NewSessionService(
	sessions ports.SessionRepository,
	signer ports.TokenSigner,
) *SessionService {
	return &SessionService{
		sessions: sessions,
		signer:   signer,
	}
}
