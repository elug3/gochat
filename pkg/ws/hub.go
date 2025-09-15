package ws

import "context"

type Hub struct {
	registerCh  chan *Client
	broadcastCh chan []byte
}

func NewHub() *Hub {
	return &Hub{
		registerCh:  make(chan *Client),
		broadcastCh: make(chan []byte),
	}
}

func (h *Hub) Register(client *Client) {
	h.registerCh <- client
}

func (h *Hub) Run(ctx context.Context) error {
	for {
		select {}
	}
}

type Room struct {
}
