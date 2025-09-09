package chatcmd

import (
	"context"
	"time"

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
	s.subc.Subscribe(events.SubjectContactsAll, func(event events.Event) error {
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		switch ev := event.(type) {
		case *events.GroupCreated:
			err = s.handler.OnGroupCreated(ctx, ev)
		case *events.GroupDeleted:
			err = s.handler.OnGroupDeleted(ctx, ev)
		case *events.MemberJoined:
			err = s.handler.OnMemberJoined(ctx, ev)
		case *events.MemberLeft:
			err = s.handler.OnMemberLeft(ctx, ev)
		}
		if err != nil {
			log.Error().Err(err).Msgf("failed to handle event: %s", event.Subject())
			return err
		}
		log.Info().Msgf("event handled: %s", event.Subject())
		return nil
	})
	<-ctx.Done()
	return nil
}
