package presence

import (
	"fmt"

	"github.com/elug3/gochat/pkg/presence/internal/errs"
	"github.com/elug3/gochat/pkg/presence/internal/model"
	"github.com/elug3/gochat/pkg/presence/internal/store"
	"github.com/elug3/gochat/pkg/presence/internal/store/redisstore"
)

type PresenceService struct {
	store store.PresenceStore
}

func NewPresenceService(opts *HttpOptions) (*PresenceService, error) {
	store, err := redisstore.NewPresenceStore(opts.RedisAddr)
	if err != nil {
		return nil, err
	}
	return &PresenceService{
		store: store,
	}, nil
}

func (s *PresenceService) GetPresence(userId int32) (*model.Presence, error) {
	p, err := s.store.GetUserPresence(userId)
	if err != nil {
		if err == errs.ErrNotFound {
			p = newEmptyPresence(userId)
		} else {
			return nil, err
		}
	}
	return p, nil
}

func (s *PresenceService) SetPresence(userId int32, state string) error {
	err := s.store.SetUserPresence(userId, state)
	if err != nil {
		return fmt.Errorf("failed to set presence: %w", err)
	}
	return nil
}

// returns a default offline presence if none is found
func newEmptyPresence(userId int32) *model.Presence {
	return &model.Presence{
		UserId:   userId,
		State:    "offline",
		LastSeen: 0,
	}
}
