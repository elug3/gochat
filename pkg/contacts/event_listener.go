package contacts

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

func NewEventListener(opts *Options) (*EventListener, error) {
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	eventSub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:   events.DurableContacts,
		AckPolicy: nats.AckExplicitPolicy,
		FilterSubjects: []string{
			events.SubjectUserRegistered,
		},
	})
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
	el.eventSub.HandleFunc(events.SubjectUserRegistered, el.handler.OnUserRegistered)

	return el.eventSub.Run(ctx)
}
