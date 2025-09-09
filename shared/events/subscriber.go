package events

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

type HandlerFn func(event Event) error

func NewSubscriber(natsUrl string) (*Subscriber, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Nats: %s: %w", natsUrl, err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}
	return &Subscriber{
		nc: nc,
		js: js,
	}, nil
}

func (s *Subscriber) Subscribe(subj string, handlerFn HandlerFn) error {
	_, err := s.js.Subscribe(">", func(msg *nats.Msg) {
		event, err := UnmarshalEvent(msg.Subject, msg.Data)
		if err != nil {
			msg.Nak()
			return
		}
		if err := handlerFn(event); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()

	}, nats.Durable("contacts-subscriber"), nats.ManualAck(), nats.BindStream("CONTACTS"))
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subj, err)
	}
	return nil
}
