package ws

import (
	"context"
	"encoding/json"

	"github.com/elug3/gochat/shared/events"
)

type EventHandler struct {
	hub *Hub
}

func NewEventHandler(hub *Hub) *EventHandler {
	return &EventHandler{
		hub: hub,
	}
}

func (h *EventHandler) OnMessageSent(ctx context.Context, subject string, data []byte) error {
	var event events.MessageSent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}
	h.hub.broadcastCh <- &BroadcastMsg{
		ChatId:    event.ChatId,
		SenderId:  int32(event.SenderId), // TODO: update event.senderId to int32
		Timestamp: event.Timestamp,
		Content:   event.Content,
	}
	return nil
}

func OnGroupJoined() {
}
