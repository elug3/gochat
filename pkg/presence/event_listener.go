package presence

import (
	"context"
	"os"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
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
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	eventSub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
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
	el.eventSub.HandleFunc(events.SubjectWebsocketConnected, el.handler.OnConnected)
	el.eventSub.HandleFunc(events.SubjectWebsocketDisconnected, el.handler.OnDisconnected)

	return el.eventSub.Run(ctx)
}

func (el *EventListener) Close() {
	if el.eventSub != nil {
		el.eventSub.Close()
	}
}
