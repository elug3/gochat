package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewPublisher(natsUrl string) (*Publisher, error) {
	nc, err := nats.Connect(
		natsUrl,
		nats.Name("publisher"),
		nats.Timeout(5*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(500*time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server %s: %w", natsUrl, err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	return &Publisher{nc: nc, js: js}, nil
}

func (p *Publisher) Close() {
	if p == nil || p.nc == nil {
		return
	}
	p.nc.Drain()
	p.nc.Close()
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
