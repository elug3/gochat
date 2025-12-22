package contacts

import (
	"context"
	"os"
	"sync"

	"github.com/elug3/gochat/pkg/contacts/internal/store/sqlite3"
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

func NewEventListener(opts *Options) (*EventListener, error) {
	contactsStore, isNewStore, err := sqlite3.NewContactsStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return nil, err
	}
	var pub *events.Publisher
	if !opts.NoEvent {
		pub, err = events.NewPublisher(opts.NatsUrl)
		if err != nil {
			return nil, err
		}
	}

	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})

	eventSub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:   events.DurableContacts,
		AckPolicy: nats.AckExplicitPolicy,
		FilterSubjects: []string{
			events.SubjectUserRegistered,
		},
	})
	if err != nil {
		return nil, err
	}
	contacts, err := NewContactsService(ServiceDeps{
		Store:      contactsStore,
		Publisher:  pub,
		IsNewStore: isNewStore,
	})
	if err != nil {
		return nil, err
	}
	h := NewEventHandler(contacts)
	return &EventListener{
		eventSub: eventSub,
		handler:  h,
		nc:       nc,
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

	el.eventSub.HandleFunc(events.SubjectUserRegistered, el.handler.OnUserRegistered)

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
