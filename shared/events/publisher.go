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
	if _, err := js.StreamInfo("CONTACTS"); err != nil {
		if err == nats.ErrStreamNotFound {
			if _, err = js.AddStream(&nats.StreamConfig{
				Name:     "CONTACTS",
				Subjects: []string{SubjectContactsAll},
				Storage:  nats.FileStorage,
			}); err != nil {
				return fmt.Errorf("failed to create CONTACTS stream: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get CONTACTS stream info: %w", err)

		}
	}
	if _, err := js.StreamInfo("MESSAGES"); err != nil {
		if err == nats.ErrStreamNotFound {
			if _, err = js.AddStream(&nats.StreamConfig{
				Name:     "MESSAGES",
				Subjects: []string{SubjectMessageAll},
				Storage:  nats.FileStorage,
			}); err != nil {
				return fmt.Errorf("failed to create MESSAGES stream: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get MESSAGES stream info: %w", err)
		}
	}
	if _, err := js.StreamInfo("WEBSOCKET"); err != nil {
		if err == nats.ErrStreamNotFound {
			if _, err = js.AddStream(&nats.StreamConfig{
				Name:     "WEBSOCKET",
				Subjects: []string{SubjectWebsocketAll},
				Storage:  nats.FileStorage,
			}); err != nil {
				return fmt.Errorf("failed to create WEBSOCKET stream: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get WEBSOCKET stream info: %w", err)
		}
	}

	return nil
}

func initConsumers(js nats.JetStreamContext) error {
	cons := map[string]struct {
		stream      string
		description string
		cfg         *nats.ConsumerConfig
		opts        []nats.JSOpt
	}{
		"CHATVIEW_CONTACTS": {
			stream:      "CONTACTS",
			description: "chatview service consumes contacts events for write model",
			cfg: &nats.ConsumerConfig{
				Durable:   "chatview",
				AckPolicy: nats.AckNonePolicy,
			},
		},
		"CHATVIEW_MESSAGES": {
			stream:      "MESSAGES",
			description: "chatview service consumes messages for write model",
			cfg: &nats.ConsumerConfig{
				Durable:   "chatview",
				AckPolicy: nats.AckNonePolicy,
			},
		},
		"MESSAGE_WEBSOCKET": {
			stream:      "WEBSOCKET",
			description: "message service consumes websocket events to send messages.",
			cfg: &nats.ConsumerConfig{
				Durable:    "message",
				AckPolicy:  nats.AckExplicitPolicy,
				MaxDeliver: 3,
			},
		},
	}

	for name, con := range cons {
		if _, err := js.ConsumerInfo(con.stream, name); err != nil {
			if err == nats.ErrConsumerNotFound {
				if _, err = js.AddConsumer(con.stream, con.cfg, con.opts...); err != nil {
					return fmt.Errorf("failed to create %s consumer on %s stream: %w", name, con.stream, err)
				}
			} else {
				return fmt.Errorf("failed to get consumer info: %w", err)
			}
		}
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
