package presence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elug3/gochat/pkg/presence/internal/model"
	"github.com/elug3/gochat/shared/events"
)

type EventHandler struct {
	presences *PresenceService
}

func NewEventHandler(presences *PresenceService) (*EventHandler, error) {
	if presences == nil {
		return nil, fmt.Errorf("presence service is required")
	}
	return &EventHandler{presences: presences}, nil
}

func (h *EventHandler) OnConnected(ctx context.Context, subject string, data []byte) error {
	var event events.WebsocketConnected
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("unmarshal event data: %w", err)
	}
	if event.IsFirst {
		err := h.presences.SetPresence(ctx, event.UserId, model.StateOnline)
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
		err := h.presences.SetPresence(ctx, event.UserId, model.StateOffline)
		if err != nil {
			return err
		}
	}
	return nil
}
