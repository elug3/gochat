package message

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/elug3/gochat/api/httpclient"
	"github.com/elug3/gochat/pkg/message/internal/store/sqlite3"
	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	deps, err := newMessageServiceDeps(opts)
	if err != nil {
		return nil, err
	}
	s, err := NewMessageService(deps)
	if err != nil {
		return nil, err
	}
	h := NewMessageHandler(s)
	server := &http.Server{
		Addr:    addr,
		Handler: h,
	}
	return server, nil
}

type EventServer struct {
	EventSub *events.Subscriber
	Messages *MessageService
	nc       *nats.Conn

	shutdownOnce sync.Once
	wg           sync.WaitGroup
	Cancel       context.CancelFunc
}

func NewEventServer(opts *Options) (*EventServer, error) {
	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	eventSub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurableMessage,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "message service consumer",
		FilterSubjects: []string{
			events.SubjectWebsocketSent,
		},
	})
	if err != nil {
		return nil, err
	}

	deps, err := newMessageServiceDeps(opts)
	if err != nil {
		return nil, err
	}
	s, err := NewMessageService(deps)
	if err != nil {
		return nil, err
	}

	return &EventServer{
		EventSub: eventSub,
		Messages: s,
		nc:       nc,
	}, nil
}

func newMessageServiceDeps(opts *Options) (ServiceDeps, error) {
	store, err := sqlite3.NewMessageStore(opts.SaveDir, opts.NoSave)
	if err != nil {
		return ServiceDeps{}, err
	}
	pub, err := events.NewPublisher(opts.NatsUrl)
	if err != nil {
		return ServiceDeps{}, err
	}
	contactsClient := httpclient.NewContactsClient(
		httpclient.WithBaseUrl(opts.ContactsServerUrl),
	)
	return ServiceDeps{
		Store:     store,
		Publisher: pub,
		Contacts:  contactsClient,
	}, nil
}

func (srv *EventServer) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	srv.Cancel = cancel
	srv.wg.Add(1)
	defer func() {
		srv.wg.Done()
		srv.Shutdown(context.Background())
	}()

	srv.EventSub.HandleFunc(events.SubjectWebsocketSent, func(ctx context.Context, subject string, data []byte) error {
		var event events.WebsocketSent
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}
		_, err := srv.Messages.Send(ctx, event.SenderId, event.ChatId, event.Content)
		if err != nil {
			return err
		}
		return nil
	})
	return srv.EventSub.Run(runCtx)
}

func (srv *EventServer) Shutdown(ctx context.Context) {
	if srv == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	srv.shutdownOnce.Do(func() {
		if srv.Cancel != nil {
			srv.Cancel()
		}
		if srv.EventSub != nil {
			_ = srv.EventSub.Close()
		}
		if srv.nc != nil && !srv.nc.IsClosed() {
			srv.nc.Close()
		}

		done := make(chan struct{})
		go func() {
			srv.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}

// func (srv *EventServer) Run(ctx context.Context) error {
// 	// TODO: handle all type of messages (http, websocket)
// 	err := srv.EventSub.PullSubscribe(ctx, func(event events.Event) error {
// 		switch ev := event.(type) {
// 		case *events.WebsocketSent:
// 			_, err := srv.Messages.Send(ctx, ev.SenderId, ev.ChatId, ev.Content)
// 			if err != nil {
// 				return err
// 			}
// 			log.Info().Msgf("Relayed websocket message from user %d to chat %d", ev.SenderId, ev.ChatId)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err := srv.EventSub.Drain(); err != nil {
// 			log.Warn().Err(err).Msg("failed to drain event subscriber")
// 		}
// 		srv.EventSub.Close()
// 	}()

// 	<-ctx.Done()
// 	return nil
// }
//
