package store

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized     = errors.New("Unauthorized")
	ErrPermissionDenied = errors.New("PermissionDenied")
	ErrNotFound         = errors.New("Notfound")
	ErrExists           = errors.New("AlreadyExists")
	ErrBadRequest       = errors.New("BadRequest")
)

type Kind string

const (
	KindUser    = "user"
	KindProfile = "profile"
	KindGroup   = "gruop"
	KindMember  = "member"
)

type Error struct {
	Kind    Kind
	Err     error
	Message string
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s", err.Message)
}

func (err *Error) Unwrap() error {
	return err.Err
}
