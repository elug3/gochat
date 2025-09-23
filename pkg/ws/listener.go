package ws

import (
	"context"

	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	sub *events.Subscriber
	h   *EventHandler
}

func NewEventListener(hub *Hub, opts *Options) (*EventListener, error) {
	sub, err := events.NewSubscriber(opts.NatsUrl)
	if err != nil {
		log.Error().Err(err).Msg("failed to create subscriber")
		return nil, err
	}

	return &EventListener{
		sub: sub,
		h:   NewEventHandler(hub),
	}, nil
}

func (el *EventListener) Run(ctx context.Context) error {
	el.sub.SubscribeStreams([]string{
		"CONTACTS",
		"MESSAGES",
	}, "WS_SERVER", func(event events.Event) error {
		switch event := event.(type) {
		case *events.MessageSent:
			return el.h.OnMessageSent(event)
		default:
			log.Warn().Msgf("unhandled event type: %T", event)
		}
		return nil
	})
	defer el.sub.Drain()

	<-ctx.Done()
	return nil
}
