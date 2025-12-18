package ws

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

func NewEventListener(hub *Hub, opts *Options) (*EventListener, error) {
	sub, err := events.NewSubscriber(opts.NatsUrl, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurableWebsocket,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "websocket service consumer",
		FilterSubjects: []string{
			events.SubjectMessageSent,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create subscriber")
		return nil, err
	}

	return &EventListener{
		eventSub: sub,
		handler:  NewEventHandler(hub),
	}, nil
}

func (el *EventListener) Run(ctx context.Context) error {
	log.Info().Msg("starting websocket event listener")
	err := el.eventSub.PullSubscribe(ctx, el.handleEvent)
	if err != nil {
		log.Error().Err(err).Msg("failed to start websocket event listener")
		return err
	}

	defer el.eventSub.Drain()
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
