package contacts

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

func NewEventListener(opts *Options) (*EventListener, error) {
	eventSub, err := events.NewSubscriber(opts.NatsUrl, events.APP_STREAM, &nats.ConsumerConfig{
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
	log.Info().Msg("starting event listener")
	err := el.eventSub.PullSubscribe(ctx, func(event events.Event) error {
		var err error
		switch ev := event.(type) {
		case *events.UserRegistered:
			err = el.handler.OnUserRegistered(ctx, ev)
		default:
			log.Warn().Msgf("unhandled event type: %T", ev)
			return nil
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

	defer func() {
		if err := el.eventSub.Drain(); err != nil {
			log.Warn().Err(err).Msg("failed to drain event subscriber")
		}
		el.eventSub.Close()
	}()

	<-ctx.Done()
	return nil
}
