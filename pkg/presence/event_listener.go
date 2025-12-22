package presence

import (
	"context"
	"os"
	"sync"

	"github.com/elug3/gochat/pkg/presence/internal/store/redisstore"
	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type EventListener struct {
	eventSub *events.Subscriber
	handler  *EventHandler
	nc       *nats.Conn

	shutdownOnce sync.Once
	wg           sync.WaitGroup
	Cancel       context.CancelFunc
}

func NewEventListener(opts *EventOptions) (*EventListener, error) {
	store, err := redisstore.NewPresenceStore(opts.RedisAddr)
	if err != nil {
		return nil, err
	}
	presences, err := NewPresenceService(store)
	if err != nil {
		return nil, err
	}
	handler, err := NewEventHandler(presences)
	if err != nil {
		return nil, err
	}
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	eventSub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurablePresence,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "presence service consumer",
		FilterSubjects: []string{
			events.SubjectWebsocketConnected,
			events.SubjectWebsocketDisconnected,
		},
	})
	if err != nil {
		return nil, err
	}

	el := EventListener{
		eventSub: eventSub,
		handler:  handler,
		nc:       nc,
	}
	return &el, nil
}

func (el *EventListener) Listen(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	el.Cancel = cancel
	el.wg.Add(1)
	defer func() {
		el.wg.Done()
		el.Shutdown(context.Background())
	}()

	el.eventSub.HandleFunc(events.SubjectWebsocketConnected, el.handler.OnConnected)
	el.eventSub.HandleFunc(events.SubjectWebsocketDisconnected, el.handler.OnDisconnected)

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
		if el.nc != nil && !el.nc.IsClosed() {
			el.nc.Close()
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
