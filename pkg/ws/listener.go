package ws

import (
	"context"
	"os"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	eventSub *events.Subscriber
	handler  *EventHandler
}

func NewEventListener(hub *Hub, opts *Options) (*EventListener, error) {
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	sub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
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
	defer el.Close()
	el.eventSub.HandleFunc(events.SubjectMessageSent, el.handler.OnMessageSent)

	return el.eventSub.Run(ctx)
}

func (el *EventListener) Close() {
	if el.eventSub != nil {
		el.eventSub.Close()
	}
}
