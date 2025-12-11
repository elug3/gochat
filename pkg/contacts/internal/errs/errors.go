package errs

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrGroupNotExists   = errors.New("group not exists")
	ErrUserNotExists    = errors.New("user not exists")
	ErrExists           = errors.New("already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrSelfContact      = errors.New("cannot add self to contacts")
)
