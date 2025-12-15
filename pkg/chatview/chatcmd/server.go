package chatcmd

import (
	"context"
	"fmt"

	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type ChatProjectionServer struct {
	subc    *events.Subscriber
	handler *ChatCommandHandler
}

func NewChatProjectionServer(opts *Options) (*ChatProjectionServer, error) {
	subc, err := events.NewSubscriber(opts.NatsUrl)
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

	err := s.subc.PullSubscribeStream(ctx, ">", "CONTACTS", "CHATCMD_CONTACTS", s.handleEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to CONTACTS stream: %w", err)
	}

	if err = s.subc.PullSubscribeStream(ctx, ">", "MESSAGES", "CHATCMD_MESSAGES", s.handleEvent); err != nil {
		return fmt.Errorf("failed to subscribe to MESSAGES stream: %w", err)
	}

	<-ctx.Done()
	log.Info().Msg("shutting down server ...")
	if err := s.subc.Drain(); err != nil {
		return err
	}
	s.subc.Close()
	log.Info().Msg("server shut down")
	return nil
}

func (s *ChatProjectionServer) handleEvent(event events.Event) error {
	ctx := context.TODO()

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
		err = fmt.Errorf("unhandled event type: %T", ev)
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to handle event")
	}
	log.Info().Msgf("handled event: %T", event)

	return err
}
