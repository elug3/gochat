package ws

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	sub     *events.Subscriber
	handler *EventHandler
}

func NewEventListener(hub *Hub, opts *Options) (*EventListener, error) {
	sub, err := events.NewSubscriber(opts.NatsUrl)
	if err != nil {
		log.Error().Err(err).Msg("failed to create subscriber")
		return nil, err
	}

	return &EventListener{
		sub:     sub,
		handler: NewEventHandler(hub),
	}, nil
}

func (el *EventListener) Run(ctx context.Context) error {
	err := el.sub.SubscribeStream(events.StreamContacts, "WS_SERVER", el.handleEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to contacts stream: %w", err)
	}
	if err = el.sub.SubscribeStream(events.StreamMessages, "WS_SERVER", el.handleEvent); err != nil {
		return fmt.Errorf("failed to subscribe to messages stream: %w", err)
	}

	defer el.sub.Drain()

	<-ctx.Done()
	return nil
}

func (el *EventListener) handleEvent(event events.Event) error {
	switch event := event.(type) {
	case *events.MessageSent:
		return el.handler.OnMessageSent(event)
	default:
		log.Warn().Msgf("unhandled event type: %T", event)
	}
	return nil
}
