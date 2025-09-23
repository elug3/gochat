package chatcmd

import (
	"context"

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
	var err error

	s.subc.SubscribeStreams([]string{
		"CONTACTS",
		"MESSAGES",
	}, "CHATVIEW", func(event events.Event) error {
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
		case *events.MessaageRead:
			err = s.handler.OnMessageRead(ctx, ev)
		default:
			log.Warn().Msgf("unknown event type: %T", event)
			return nil
		}
		if err != nil {
			log.Error().Err(err).Msgf("failed to handle event: %T", event)
			return err
		}
		log.Info().Msgf("event handled: %s", event.Subject())
		return nil
	})
	if err != nil {
		return err
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
