package ws

import "github.com/elug3/gochat/shared/events"

type EventHandler struct {
	subc *events.Subscriber
	hub  *Hub
}

func NewEventHandler(hub *Hub) *EventHandler {
	return &EventHandler{}
}

func (h *EventHandler) Run() error {
	return nil
}

func OnGroupJoined() {
}
