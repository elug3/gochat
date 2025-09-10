package events

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewPublisher(natsUrl string) (*Publisher, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server %s: %w", natsUrl, err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	if err := initStream(js); err != nil {
		return nil, err
	}
	if err := initConsumers(js); err != nil {
		return nil, err
	}

	return &Publisher{
		nc: nc,
		js: js,
	}, nil
}

func initStream(js nats.JetStreamContext) error {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "CONTACTS",
		Subjects: []string{SubjectContactsAll},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to create CONTACTS stream: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "MESSAGES",
		Subjects: []string{SubjectMessageAll},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to create MESSAGES stream: %w", err)
	}
	return nil
}

func initConsumers(js nats.JetStreamContext) error {
	_, err := js.AddConsumer("CONTACTS", &nats.ConsumerConfig{
		Durable:       "CHATVIEW_MESSAGE",
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create CHATVIEW_MESSAGES consumer: %w", err)
	}
	if _, err = js.AddConsumer("CONTACTS", &nats.ConsumerConfig{
		Durable:       "CHATVIEW_CONTACTS",
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	}); err != nil {
		return fmt.Errorf("failed to create CONTACTS consumer: %w", err)
	}

	return nil
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
