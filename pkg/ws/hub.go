package ws

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coder/websocket"
	"github.com/elug3/gochat/shared/events"
	"github.com/rs/zerolog/log"
)

type Hub struct {
	clientIdSeq int
	eventPub    events.Publisher

	registerCh chan *RegisterMsg

	// unregisterCh is used to unregister a client from the hub
	unregisterCh chan *UnregisterMsg
	// subscribeCh is used to subscribe a user to a chat.
	// User must register before subscribing. If not registered, the subscribe request is ignored.
	subscribeCh chan *SubscribeMsg
	// unsubscribeCh is used to unsubscribe a user from a chat
	unsubscribeCh chan *int
	// broadcastCh is used to broadcast message to clients in a chat
	broadcastCh chan *BroadcastMsg
	// clientMsgCh is used to receive message from websocket clients
	clientMsgCh chan *ClientMsg

	users map[int32]*User
	chats map[int]*Chat
}

func NewHub(eventPub *events.Publisher) *Hub {
	return &Hub{
		eventPub:     *eventPub,
		registerCh:   make(chan *RegisterMsg),
		unregisterCh: make(chan *UnregisterMsg),

		subscribeCh:   make(chan *SubscribeMsg),
		unsubscribeCh: make(chan *int),
		broadcastCh:   make(chan *BroadcastMsg),
		clientMsgCh:   make(chan *ClientMsg),

		users: make(map[int32]*User),
		chats: make(map[int]*Chat),
	}
}

type User struct {
	Id          int32
	Clients     map[int]*Client
	BroadcastCh chan *BroadcastMsg
}

type Chat struct {
	Users       map[int32]*User
	BroadcastCh chan *BroadcastMsg
}

type RegisterMsg struct {
	UserId int32
	Conn   *websocket.Conn
}

// UnregisterMsg is used to unregistering User from the hub.
type UnregisterMsg struct {
	UserId int32
	// Specifies the target for unregistering. If 0, it specifies all of the user's clients.
	ClientId int
}

type SubscribeMsg struct {
	ChatId int
	UserId int32
}

type BroadcastMsg struct {
	ChatId    int
	SenderId  int32
	TimeStamp int64
	Data      []byte
}

// Register is shorthand for adding a client to the hub's register channel.
func (h *Hub) Register(userId int32, conn *websocket.Conn) {
	h.registerCh <- &RegisterMsg{UserId: userId, Conn: conn}
}

func (h *Hub) Run(ctx context.Context) error {
	for {
		select {
		case msg := <-h.registerCh:
			var user *User
			var ok bool

			clientId := h.incClientId()
			client := &Client{
				Id:     clientId,
				conn:   msg.Conn,
				UserId: msg.UserId,
			}

			// create user if not exists
			if user, ok = h.users[msg.UserId]; !ok {
				user = &User{Id: msg.UserId, Clients: make(map[int]*Client)}
				h.users[msg.UserId] = user
			}
			// add client to user
			user.Clients[clientId] = client

			// start reading messages from client
			go func() {
				ch := client.ReadChan(ctx)
				for msg := range ch {
					h.clientMsgCh <- msg
				}
				if !client.closed {
					h.unregisterCh <- &UnregisterMsg{UserId: client.UserId, ClientId: clientId}
				}
			}()
			log.Info().Msgf("user %d registered with client %d", msg.UserId, clientId)

		case msg := <-h.unregisterCh:
			var user *User
			var ok bool
			// ignore unregistered user clients
			if user, ok = h.users[msg.UserId]; ok {
				if msg.ClientId == 0 {
					length := len(user.Clients)
					// all user clients
					for i, c := range user.Clients {
						c.Close()
						delete(user.Clients, i)
					}
					log.Info().Msgf("user: %d unregistered all clients (%d)", msg.UserId, length)

				} else if c, ok := user.Clients[msg.ClientId]; ok {
					c.Close()
					delete(user.Clients, msg.ClientId)
					log.Info().Msgf("user: %d unregistered client %d", msg.UserId, msg.ClientId)
				}
				if len(user.Clients) == 0 {
					delete(h.users, msg.UserId)
				}
			}
		case msg := <-h.subscribeCh:
			var (
				user *User
				chat *Chat
				ok   bool
			)
			// ignore unregistered user
			if user, ok = h.users[msg.UserId]; ok {
				// create chat if not exists
				if chat, ok = h.chats[msg.ChatId]; !ok {
					chat = newChat()
					h.chats[msg.ChatId] = chat

				}
				// mapping user to chat
				chat.Users[msg.UserId] = user
				log.Info().Msgf("user %d subscribed to chat %d", msg.UserId, msg.ChatId)

			}

			// TODO: use broadcast channel
		case msg := <-h.clientMsgCh:
			if chat, ok := h.chats[msg.ChatId]; ok {
				// only broadcast if the sender is part of the chat
				if _, ok = chat.Users[msg.SenderId]; ok {
					// broadcast to all users in the chat
					for _, user := range chat.Users {
						for _, client := range user.Clients {
							data, err := json.Marshal(msg)
							if err != nil {
								log.Warn().Err(err).Msgf("failed to marshal client message for chat %d", msg.ChatId)
								continue
							}
							if err := client.conn.Write(ctx, websocket.MessageText, data); err != nil {
								log.Warn().Err(err).Msgf("failed to write message to user %d client %d", user.Id, client.Id)
								continue
							}
							log.Info().Msgf("websocket message broadcasted to '%d'", msg.ChatId)
						}
					}
					// publish event
					if err := h.eventPub.Publish(&events.WebsocketSent{
						ChatId:    msg.ChatId,
						SenderId:  msg.SenderId,
						Content:   msg.Content,
						TimeStamp: time.Now().Unix(),
					}); err != nil {
						log.Error().Err(err).Msgf("failed to publish event: %v", err)
					}

				}
			}
			// TODO: use broadcast channel
		case msg := <-h.broadcastCh:
			if chat, ok := h.chats[msg.ChatId]; ok {
				for _, user := range chat.Users {
					for _, client := range user.Clients {
						data, err := json.Marshal(msg)
						if err != nil {
							log.Warn().Err(err).Msgf("failed to marshal broadcast message for chat %d", msg.ChatId)
							continue
						}
						if err := client.conn.Write(ctx, websocket.MessageText, data); err != nil {
							log.Warn().Err(err).Msgf("failed to write message to user %d client %d", user.Id, client.Id)
							continue
						}
						log.Info().Msgf("event message broadcasted to '%d'", msg.ChatId)
					}
				}
			}
		}
	}
}

func newChat() *Chat {
	return &Chat{
		Users: make(map[int32]*User),
	}
}

func (u *User) Run(ctx context.Context) error {
	for {
		select {
		case msg := <-u.BroadcastCh:
			for _, client := range u.Clients {
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *Hub) incClientId() int {
	h.clientIdSeq += 1
	return h.clientIdSeq
}

// marge function is fan-in for merging multiple channels into one channel
func merge[T any](cs []<-chan T) <-chan T {
	out := make(chan T)

	output := func(c <-chan T) {
		for x := range c {
			out <- x
		}
	}

	for _, c := range cs {
		go output(c)
	}

	return out
}
