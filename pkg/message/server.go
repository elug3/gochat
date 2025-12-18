package message

import (
	"context"
	"net"
	"net/http"

	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func NewHttpServer(opts *Options) (*http.Server, error) {
	addr := net.JoinHostPort(opts.Host, opts.Port)
	s, err := NewMessageService(opts)
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
}

func NewEventServer(opts *Options) (*EventServer, error) {
	eventSub, err := events.NewSubscriber(opts.NatsUrl, events.APP_STREAM, &nats.ConsumerConfig{
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

	s, err := NewMessageService(opts)
	if err != nil {
		return nil, err
	}

	return &EventServer{
		EventSub: eventSub,
		Messages: s,
	}, nil
}

func (srv *EventServer) Run(ctx context.Context) error {
	// TODO: handle all type of messages (http, websocket)
	err := srv.EventSub.PullSubscribe(ctx, func(event events.Event) error {
		switch ev := event.(type) {
		case *events.WebsocketSent:
			_, err := srv.Messages.Send(ctx, ev.SenderId, ev.ChatId, ev.Content)
			if err != nil {
				return err
			}
			log.Info().Msgf("Relayed websocket message from user %d to chat %d", ev.SenderId, ev.ChatId)
		}
		return nil
	})
	if err != nil {
		return err
	}

	defer func() {
		if err := srv.EventSub.Drain(); err != nil {
			log.Warn().Err(err).Msg("failed to drain event subscriber")
		}
		srv.EventSub.Close()
	}()

	<-ctx.Done()
	return nil
}
