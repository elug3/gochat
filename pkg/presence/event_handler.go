package presence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elug3/gochat/pkg/presence/internal/model"
	"github.com/elug3/gochat/pkg/presence/internal/store/redisstore"
	"github.com/elug3/gochat/shared/events"
)

type EventHandler struct {
	store *redisstore.PresenceStore
}

func NewEventHandler(opts *EventOptions) (*EventHandler, error) {
	store, err := redisstore.NewPresenceStore(opts.RedisAddr)
	if err != nil {
		return nil, err
	}
	return &EventHandler{store: store}, nil
}

func (h *EventHandler) OnConnected(ctx context.Context, subject string, data []byte) error {
	var event events.WebsocketConnected
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal event data: %w", err)
	}
	if event.IsFirst {
		err := h.store.SetUserPresence(ctx, event.UserId, model.StateOnline)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *EventHandler) OnDisconnected(ctx context.Context, subject string, data []byte) error {
	var event events.WebsocketDisconnected
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal event data: %w", err)
	}
	if event.IsLast {
		err := h.store.SetUserPresence(ctx, event.UserId, model.StateOffline)
		if err != nil {
			return err
		}
	}
	return nil
}
