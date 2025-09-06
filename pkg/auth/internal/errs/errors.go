package errs

import "errors"

var (
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrNotFound              = errors.New("not found")
	ErrExists                = errors.New("already exists")
)
