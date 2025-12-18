package presence

import (
	"context"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
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
	eventSub, err := events.NewSubscriber(opts.NatsUrl, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurablePresence,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "presence service consumer",
		FilterSubjects: []string{
			events.SubjectWebsocketConnected,
			events.SubjectWebsocketDisconnected,
		},
	})
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
	err := el.eventSub.PullSubscribe(ctx, func(event events.Event) error {
		switch ev := event.(type) {
		case *events.WebsocketConnected:
			if err := el.handler.OnConnected(ev); err != nil {
				log.Error().Err(err).Msgf("error handling event: %T", event)
				return err
			}
		case *events.WebsocketDisconnected:
			if err := el.handler.OnDisconnected(ev); err != nil {
				log.Error().Err(err).Msgf("error handling event: %T", event)
				return err
			}

		default:
			log.Warn().Msgf("unhandled event type: %T", event)
			return nil
		}

		log.Info().Msgf("handled event: %T", event)
		return nil
	})
	if err != nil {
		return err
	}
	log.Info().Msg("listening for events...")

	defer func() {
		if err := el.eventSub.Drain(); err != nil {
			log.Warn().Err(err).Msg("failed to drain event subscriber")
		}
		el.eventSub.Close()
	}()

	<-ctx.Done()
	return nil
}
