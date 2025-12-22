package contacts

import (
	"context"
	"encoding/json"
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

func (h *EventHandler) OnUserRegistered(ctx context.Context, subject string, data []byte) error {
	var event events.UserRegistered
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	_, err := h.Contacts.CreateProfile(ctx, event.UserId, event.Username)
	if err != nil {
		if errors.Is(err, errs.ErrExists) {
			// CreateProfile is idempotent
			return nil
		}
		return err
	}
	return nil
}
