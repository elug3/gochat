package ws

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
)

type Client struct {
	Id     int
	UserId int32
	conn   *websocket.Conn
	closed bool
	cancel context.CancelFunc
}

type ClientMsg struct {
	ChatId   int    `json:"chat_id"`
	SenderId int32  `json:"sender_id"`
	Content  string `json:"content"`
}

func (c *Client) ReadChan(ctx context.Context) <-chan *ClientMsg {
	out := make(chan *ClientMsg)
	ctx, c.cancel = context.WithCancel(ctx)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return

			default:
				var clientMsg ClientMsg
				typ, msg, err := c.conn.Read(ctx)
				if err != nil {
					return
				}
				if typ == websocket.MessageText {
					if err := json.Unmarshal(msg, &clientMsg); err != nil {
						log.Error().Err(err).Msg("failed to unmarshal message")
						continue
					}
					clientMsg.SenderId = c.UserId
					out <- &clientMsg
				}
			}
		}
	}()

	return out
}

func (c *Client) WriteChan(ctx context.Context, ch chan *BroadcastMsg) {

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, []byte(msg.Content)); err != nil {
				log.Error().Err(err).Msg("failed to write message to client")
				return
			}
		}
	}
}

func (c *Client) Close() {
	if c.closed {
		panic("client already closed")
	}
	if c.cancel != nil {
		c.cancel()
	}
	c.closed = true
}
