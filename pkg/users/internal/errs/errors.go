package errs

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrExists                = errors.New("already exists")
	ErrInvalid               = errors.New("invalid input")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrPermissionDenied      = errors.New("permission denied")
	ErrTimeout               = errors.New("operation timed out")
	ErrAuthenticationFailure = errors.New("authentication failure")
)
