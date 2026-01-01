package errs

import "errors"

var (
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrSessionExpired        = errors.New("session expired")
	ErrNotFound              = errors.New("not found")
	ErrExists                = errors.New("already exists")
	ErrConstraint            = errors.New("constraint violation")
)

type PermissionError struct{}
