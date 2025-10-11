package store

import "github.com/elug3/gochat/pkg/presence/internal/model"

type PresenceStore interface {
	SetUserPresence(userId int32, state string) error
	GetUserPresence(userId int32) (*model.Presence, error)
}
