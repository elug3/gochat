package event

import "github.com/nats-io/nats.go"

type EventPublisher struct {
	nc *nats.Conn
}

func NewEventPublisher(natsUrl string) (*EventPublisher, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, err
	}

	return &EventPublisher{
		nc: nc,
	}, nil
}

func (p *EventPublisher) Publish(event Event) error {
}
