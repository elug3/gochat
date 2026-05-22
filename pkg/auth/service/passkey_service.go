package service

import "github.com/elug3/gochat/pkg/auth/ports"

// PasskeyService encapsulates WebAuthn/passkey workflows.
type PasskeyService struct {
	passkeys ports.PasskeyProvider
	users    ports.UserRepository
}

func NewPasskeyService(
	passkeys ports.PasskeyProvider,
	users ports.UserRepository,
) *PasskeyService {
	return &PasskeyService{
		passkeys: passkeys,
		users:    users,
	}
}
