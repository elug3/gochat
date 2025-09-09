package events

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewPublisher(natsUrl string) (*Publisher, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	if _, err = js.AddStream(&nats.StreamConfig{
		Name: "CONTACTS",
		Subjects: []string{
			"contacts.group.*",
			"contacts.member.*",
		},
	}); err != nil {
		return nil, err
	}

	return &Publisher{
		nc: nc,
		js: js,
	}, nil
}

func (p *Publisher) Publish(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := p.js.Publish(event.Subject(), data); err != nil {
		return err
	}
	return nil
}
