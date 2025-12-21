package chatcmd

import (
	"context"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type ChatProjectionServer struct {
	subc    *events.Subscriber
	handler *ChatCommandHandler
}

func NewChatProjectionServer(opts *Options) (*ChatProjectionServer, error) {
	subc, err := events.NewSubscriber(opts.NatsUrl, events.APP_STREAM, &nats.ConsumerConfig{
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
	return &ChatProjectionServer{
		subc:    subc,
		handler: handler,
	}, nil
}

func (s *ChatProjectionServer) Run(ctx context.Context) error {
	log.Info().Msg("server started ...")

}

func (s *ChatProjectionServer) handleEvent(ctx context.Context, event events.Event) error {
	var err error
	switch ev := event.(type) {
	case *events.GroupCreated:
		err = s.handler.OnGroupCreated(ctx, ev)
	case *events.GroupDeleted:
		err = s.handler.OnGroupDeleted(ctx, ev)
	case *events.MemberJoined:
		err = s.handler.OnMemberJoined(ctx, ev)
	case *events.MemberLeft:
		err = s.handler.OnMemberLeft(ctx, ev)
	case *events.MessageSent:
		err = s.handler.OnMessageSent(ctx, ev)
	case *events.MessageRead:
		err = s.handler.OnMessageRead(ctx, ev)
	case *events.ContactsReset:
		err = s.handler.OnContactsReset(ctx, ev)
	default:
		log.Warn().Msgf("unhandled event type: %T", ev)
		return nil
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to handle event")
		return err
	}
	log.Info().Msgf("handled event: %T", event)
	return nil
}
