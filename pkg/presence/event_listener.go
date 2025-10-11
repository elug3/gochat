package presence

import (
	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	eventSub *events.Subscriber
	handler  *EventHandler
}

func NewEventListener(opts *EventOptions) (*EventListener, error) {
	eventSub, err := events.NewSubscriber(opts.NatsUrl)
	if err != nil {
		return nil, err
	}

	el := EventListener{
		eventSub: eventSub,
	}
	return &el, nil
}

func (el *EventListener) Listen() error {
	err := el.eventSub.SubscribeStream(events.StreamWebsocket, "PRESENCE", func(event events.Event) error {
		var err error

		switch ev := event.(type) {
		case *events.WebsocketConnected:
			err = el.handler.OnConnected(ev)
		case *events.WebsocketDisconnected:
			err = el.handler.OnDisconnected(ev)

		default:
			log.Warn().Msgf("unhandled event type: %T", ev)
		}
		return err
	})

	if err != nil {
		return err
	}
	return nil
}
