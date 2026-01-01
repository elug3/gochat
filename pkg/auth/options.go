package auth

import (
	"time"
)

type Options struct {
	Host string
	Port string

	// Path to RSA private key in PEM format
	JWTKeyPath         string
	UseTemporaryJWTKey bool

	SaveDir  string
	InMemory bool

	// debug disables recovery middleware
	debug bool

	LogLevel  string
	LogFormat string

	NatsUrl  string
	NoEvents bool

	RPDisplayName string
	RPID          string
	RPOrigins     []string
}
type CreateApiTokenOptions struct {
	Name     string
	Scope    []string
	ExpireIn time.Duration
}
