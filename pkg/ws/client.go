package ws

import "github.com/coder/websocket"

type Client struct {
	UserId int32
	conn   *websocket.Conn
}

func NewClient(userId int32, conn *websocket.Conn) *Client {
	return &Client{
		UserId: userId,
		conn:   conn,
	}
}
