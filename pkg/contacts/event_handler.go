package contacts

import (
	"context"

	"github.com/elug3/gochat/shared/events"
)

type EventHandler struct {
	Contacts *ContactsService
}

func NewEventHandler(contacts *ContactsService) *EventHandler {
	return &EventHandler{
		Contacts: contacts,
	}
}

func (h *EventHandler) OnUserRegistered(ctx context.Context, ev *events.UserRegistered) error {
	_, err := h.Contacts.CreateProfile(ctx, ev.UserId, ev.Username)
	return err
}
