package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Subscriber struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	stream  string
	durable string
}

type HandlerFn func(event Event) error

func NewSubscriber(natsUrl string, stream string, cfg *nats.ConsumerConfig) (*Subscriber, error) {
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Nats: %s: %w", natsUrl, err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// if not exists, create consumer
	if _, err = js.ConsumerInfo(stream, cfg.Durable); err != nil {
		if !errors.Is(err, nats.ErrConsumerNotFound) {
			nc.Close()
			return nil, fmt.Errorf("failed to get stream info: %w", err)
		}

		// create consumer
		if _, err = js.AddConsumer(stream, cfg); err != nil {
			if _, infoErr := js.ConsumerInfo(stream, cfg.Durable); infoErr != nil {
				nc.Close()
				return nil, fmt.Errorf("failed to create consumer %s: %w", cfg.Durable, err)
			}
		}
	}

	return &Subscriber{
		nc:      nc,
		js:      js,
		stream:  stream,
		durable: cfg.Durable,
	}, nil
}

func (s *Subscriber) PullSubscribe(ctx context.Context, handler HandlerFn) error {
	sub, err := s.js.PullSubscribe(
		"", // subject ignored when binding to a durble consumer
		s.durable,
		nats.Bind(s.stream, s.durable),
	)
	if err != nil {
		return err
	}
	go func(ctx context.Context) {
		defer func() {
			if err := sub.Unsubscribe(); err != nil {
				log.Debug().Err(err).Msg("failed to unsubscribe nats subscription")
			}
		}()

		for ctx.Err() == nil {
			msgs, err := sub.Fetch(
				10,
				nats.MaxWait(5*time.Second),
				nats.Context(ctx),
			)
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				}
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}

				log.Warn().Msgf("failed to fetch messages: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(3 * time.Second):
				}
				continue
			}

			for _, msg := range msgs {
				if ctx.Err() != nil {
					return
				}

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
					msg.NakWithDelay(3 * time.Second)
					continue
				}
				if err = msg.Ack(); err != nil {
					log.Warn().Msgf("failed to ack message: %v", err)
					continue
				}
			}
		}
	}(ctx)
	return nil
}

func (s *Subscriber) Close() {
	s.nc.Close()
}

func (s *Subscriber) Drain() error {
	return s.nc.Drain()
}
