package message

import (
	"context"
	"net"
	"net/http"

	"github.com/elug3/gochat/shared/events"
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
	eventSub, err := events.NewSubscriber(opts.NatsUrl)
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
	srv.EventSub.SubscribeStream("WEBSOCKET", "messages", func(event events.Event) error {
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

	<-ctx.Done()
	return nil
}
