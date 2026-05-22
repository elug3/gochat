package service

import "github.com/elug3/gochat/pkg/auth/ports"

// TokenRuntimeService owns token issuing and validation concerns.
type TokenRuntimeService struct {
	signer ports.TokenSigner
}

func NewTokenRuntimeService(
	signer ports.TokenSigner,
) *TokenRuntimeService {
	return &TokenRuntimeService{
		signer: signer,
	}
}
