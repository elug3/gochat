package presence

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

func NewEventListener(opts *EventOptions) (*EventListener, error) {
	handler, err := NewEventHandler(opts)
	if err != nil {
		return nil, err
	}
	eventSub, err := events.NewSubscriber(opts.NatsUrl)
	if err != nil {
		return nil, err
	}

	el := EventListener{
		eventSub: eventSub,
		handler:  handler,
	}
	return &el, nil
}

func (el *EventListener) Listen(ctx context.Context) error {
	err := el.eventSub.SubscribeStream(events.StreamWebsocket, "PRESENCE", func(event events.Event) error {
		var err error

		switch ev := event.(type) {
		case *events.WebsocketConnected:
			err = el.handler.OnConnected(ev)
		case *events.WebsocketDisconnected:
			err = el.handler.OnDisconnected(ev)

		default:
			err = fmt.Errorf("unhandled event type: %T", ev)
		}
		if err == nil {
			log.Info().Msgf("handled event: %T", event)
		} else {
			log.Error().Err(err).Msgf("error handling event: %T", event)
		}
		return err
	})
	if err != nil {
		return err
	}
	log.Info().Msg("listening for events...")

	defer el.eventSub.Close()

	<-ctx.Done()
	return nil
}
