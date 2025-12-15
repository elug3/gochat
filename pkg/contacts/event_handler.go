package contacts

import (
	"context"
	"errors"

	"github.com/elug3/gochat/pkg/contacts/internal/errs"
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
	if err != nil {
		if errors.Is(err, errs.ErrExists) {
			// CreateProfile is idempotent
			return nil
		}
		return err
	}
	return nil
}
