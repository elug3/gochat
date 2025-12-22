package chatcmd

import (
	"context"
	"os"
	"sync"

	redisstore "github.com/elug3/gochat/pkg/chatview/chatcmd/internal/store/redis-store"
	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ChatProjectionServer struct {
	sub     *events.Subscriber
	handler *ChatCommandHandler
	nc      *nats.Conn

	shutdownOnce sync.Once
	wg           sync.WaitGroup
	Cancel       context.CancelFunc
}

func NewChatProjectionServer(opts *Options) (*ChatProjectionServer, error) {

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, err
	}

	sub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurableChatView,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "chatview service consumer",
		FilterSubjects: []string{
			events.SubjectContactsGroupAll,
			events.SubjectContactsMemberAll,
			events.SubjectMessageAll,
		},
	})
	if err != nil {
		return nil, err
	}
	store, err := redisstore.NewChatStore(opts.DatabaseUrl)
	if err != nil {
		return nil, err
	}
	handler, err := NewChatCommandHandler(store)
	if err != nil {
		return nil, err
	}
	registerHandlers(sub, handler)
	return &ChatProjectionServer{
		sub:     sub,
		handler: handler,
		nc:      nc,
	}, nil
}

func (s *ChatProjectionServer) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	s.Cancel = cancel
	s.wg.Add(1)
	defer func() {
		s.wg.Done()
		s.Shutdown(context.Background())
	}()

	return s.sub.Run(runCtx)
}

func registerHandlers(sub *events.Subscriber, h *ChatCommandHandler) {
	sub.HandleFunc(events.SubjectContactsReset, h.OnContactsReset)
	sub.HandleFunc(events.SubjectGroupCreated, h.OnGroupCreated)
	sub.HandleFunc(events.SubjectGroupDeleted, h.OnGroupDeleted)
	sub.HandleFunc(events.SubjectMemberJoined, h.OnMemberJoined)
	sub.HandleFunc(events.SubjectMemberLeft, h.OnMemberLeft)
	sub.HandleFunc(events.SubjectMessageSent, h.OnMessageSent)
	sub.HandleFunc(events.SubjectMessageRead, h.OnMessageRead)
}

func (s *ChatProjectionServer) Shutdown(ctx context.Context) {
	if s == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	s.shutdownOnce.Do(func() {
		if s.Cancel != nil {
			s.Cancel()
		}
		if s.sub != nil {
			_ = s.sub.Close()
		}
		if s.nc != nil && !s.nc.IsClosed() {
			s.nc.Close()
		}

		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}
