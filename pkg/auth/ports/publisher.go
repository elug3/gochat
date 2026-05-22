package ports

import "context"

// EventPublisher abstracts event transport implementations.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

// TokenSigner abstracts JWT/token signing implementations.
type TokenSigner interface {
	SignAccessToken(claims any) (string, error)
	SignRefreshToken(claims any) (string, error)
}

// PasskeyProvider abstracts WebAuthn implementations.
type PasskeyProvider interface {
	BeginRegistration(ctx context.Context, userID int32) (any, error)
	FinishRegistration(ctx context.Context, userID int32, response any) error
}
