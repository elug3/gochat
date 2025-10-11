package presence

import (
	"fmt"

	"github.com/elug3/gochat/pkg/presence/internal/model"
	"github.com/elug3/gochat/pkg/presence/internal/store"
	"github.com/elug3/gochat/shared/events"
)

type EventHandler struct {
	store store.PresenceStore
}

func (h *EventHandler) OnConnected(ev *events.WebsocketConnected) error {
	if ev == nil {
		return fmt.Errorf("event is nil")
	}
	if ev.IsFirst {
		err := h.store.SetUserPresence(ev.UserId, model.StateOnline)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *EventHandler) OnDisconnected(ev *events.WebsocketDisconnected) error {
	if ev == nil {
		return fmt.Errorf("event is nil")
	}
	if ev.IsLast {
		err := h.store.SetUserPresence(ev.UserId, model.StateOffline)
		if err != nil {
			return err
		}
	}
	return nil
}
