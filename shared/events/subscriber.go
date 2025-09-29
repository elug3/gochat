package events

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
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

func (s *Subscriber) SubscribeStream(stream, durable string, handler HandlerFn) error {
	_, err := s.js.Subscribe(">", func(msg *nats.Msg) {
		event, err := UnmarshalEvent(msg.Subject, msg.Data)
		if err != nil {
			fmt.Printf("failed to unmarshal event: %v\n", err)
			// TODO: add unhandled message queue
			if err = msg.Ack(); err != nil {
				fmt.Printf("failed to ack message: %v\n", err)
			}
			return
		}
		if err := handler(event); err != nil {
			fmt.Printf("handler error: %v\n", err)
			msg.NakWithDelay(time.Second * 3)
			return
		}
		if err := msg.Ack(); err != nil {
			fmt.Printf("failed to ack message: %v\n", err)
			return
		}
	}, nats.Durable(durable), nats.ManualAck(), nats.BindStream(stream))
	if err != nil {
		return err
	}
	return nil
}

func (s *Subscriber) PullSubscribeStream(ctx context.Context, subject, stream, durable string, handler HandlerFn) error {
	sub, err := s.js.PullSubscribe(subject, durable, nats.BindStream(stream))
	if err != nil {
		return err
	}
	go func() {
		for {
			msgs, err := sub.Fetch(10, nats.MaxWait(time.Second*5))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				log.Warn().Msgf("failed to fetch messages: %v", err)
				time.Sleep(time.Second * 3)
				continue
			}
			for _, msg := range msgs {
				event, err := UnmarshalEvent(msg.Subject, msg.Data)
				if err != nil {
					log.Warn().Msgf("failed to unmarshal event: %v", err)
					if err = msg.Term(); err != nil {
						log.Warn().Msgf("failed to term message: %v", err)
					}
					continue
				}
				if err = handler(event); err != nil {
					log.Warn().Msgf("handler error: %v", err)
					msg.NakWithDelay(time.Second * 3)
					continue
				}
				if err = msg.Ack(); err != nil {
					log.Warn().Msgf("failed to ack message: %v", err)
					continue
				}
			}
		}
	}()
	return nil
}

func (s *Subscriber) Close() {
	s.nc.Close()
}

func (s *Subscriber) Drain() error {
	return s.nc.Drain()
}
