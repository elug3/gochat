package contacts

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	eventSub *events.Subscriber
	handler  *EventHandler
}

func NewEventListener(opts *Options) (*EventListener, error) {
	eventSub, err := events.NewSubscriber(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	contacts, err := NewContactsService(opts)
	if err != nil {
		return nil, err
	}
	h := NewEventHandler(contacts)
	return &EventListener{
		eventSub: eventSub,
		handler:  h,
	}, nil
}

func (el *EventListener) Run(ctx context.Context) error {
	log.Info().Msg("starting event listener")
	err := el.eventSub.SubscribeStream("AUTH", "CONTACTS_AUTH", func(event events.Event) error {
		var err error
		switch ev := event.(type) {
		case *events.UserRegistered:
			err = el.handler.OnUserRegistered(ctx, ev)
		default:
			err = fmt.Errorf("unhandled event type: %T", ev)
		}

		if err != nil {
			log.Error().Err(err).Msg("failed to handle event")
			return err
		}

		log.Info().Msgf("handled event: %s", event.Subject())
		return nil
	})
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
