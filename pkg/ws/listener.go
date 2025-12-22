package ws

import (
	"context"
	"os"
	"sync"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventListener struct {
	eventSub *events.Subscriber
	handler  *EventHandler

	shutdownOnce sync.Once
	wg           sync.WaitGroup
	Cancel       context.CancelFunc
}

func NewEventListener(hub *Hub, opts *Options) (*EventListener, error) {
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	sub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurableWebsocket,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "websocket service consumer",
		FilterSubjects: []string{
			events.SubjectMessageSent,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create subscriber")
		return nil, err
	}

	return &EventListener{
		eventSub: sub,
		handler:  NewEventHandler(hub),
	}, nil
}

func (el *EventListener) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	el.Cancel = cancel
	el.wg.Add(1)
	defer func() {
		el.wg.Done()
		el.Shutdown(context.Background())
	}()

	el.eventSub.HandleFunc(events.SubjectMessageSent, el.handler.OnMessageSent)

	return el.eventSub.Run(runCtx)
}

func (el *EventListener) Shutdown(ctx context.Context) {
	if el == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	el.shutdownOnce.Do(func() {
		if el.Cancel != nil {
			el.Cancel()
		}
		if el.eventSub != nil {
			_ = el.eventSub.Close()
		}

		done := make(chan struct{})
		go func() {
			el.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}
