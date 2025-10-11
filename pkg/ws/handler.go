package ws

import "github.com/elug3/gochat/shared/events"

type EventHandler struct {
	hub *Hub
}

func NewEventHandler(hub *Hub) *EventHandler {
	return &EventHandler{
		hub: hub,
	}
}

func (h *EventHandler) OnMessageSent(ev *events.MessageSent) error {
	h.hub.broadcastCh <- &BroadcastMsg{
		ChatId:    ev.ChatId,
		SenderId:  int32(ev.SenderId), // TODO: update event.senderId to int32
		TimeStamp: ev.TimeStamp,
		Content:   ev.Content,
	}
	return nil
}

func OnGroupJoined() {
}
