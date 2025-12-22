package chatcmd

import (
	"context"
	"os"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ChatProjectionServer struct {
	sub     *events.Subscriber
	handler *ChatCommandHandler
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
	handler, err := NewChatCommandHandler(opts)
	if err != nil {
		return nil, err
	}
	registerHandlers(sub, handler)
	return &ChatProjectionServer{
		sub:     sub,
		handler: handler,
	}, nil
}

func (s *ChatProjectionServer) Run(ctx context.Context) error {
	return s.sub.Run(ctx)
}

func registerHandlers(sub *events.Subscriber, h *ChatCommandHandler) {
	sub.HandleFunc(events.SubjectContactsReset, h.OnContactsReset)
	sub.HandleFunc(events.SubjectGroupCreated, h.OnGroupCreated)
	sub.HandleFunc(events.SubjectGroupDeleted, h.OnGroupDeleted)
	sub.HandleFunc(events.SubjectMemberJoined, h.OnMemberJoined)
	sub.HandleFunc(events.SubjectMemberLeft, h.OnMemberLeft)
	sub.HandleFunc(events.SubjectMessageSent, h.OnMessageSent)
}
