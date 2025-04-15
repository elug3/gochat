package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

type Subscriber struct {
	msgs      chan []byte
	closeSlow func()
}

type ChatServer struct {
	subscriberMesageBuffer int
	publishLimiter         *rate.Limiter

	logf func(f string, v ...interface{})

	serveMux http.ServeMux

	subscriberMu sync.Mutex
	subscribers  map[*Subscriber]struct{}
}

func NewChatServer() *ChatServer {
	cs := &ChatServer{
		subscriberMesageBuffer: 16,
		logf:                   log.Printf,
		subscribers:            make(map[*Subscriber]struct{}),
		publishLimiter:         rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}
	cs.serveMux.HandleFunc("/subscribe", cs.subscribeHandler)
	return cs
}

func (cs *ChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cs.serveMux.ServeHTTP(w, r)
}

func (cs *ChatServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := cs.subscribe(w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		cs.logf("%v", err)
		return
	}
}

func (cs *ChatServer) subscribe(w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var c *websocket.Conn
	var closed bool

	sb := &Subscriber{
		msgs: make(chan []byte, cs.subscriberMesageBuffer),
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if c != nil {
				c.Close(websocket.StatusPolicyViolation, "connection too slow to leep up with messages")
			}
		},
	}
	cs.addSubscriber(sb)
	defer cs.deleteSubscriber(sb)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	c = conn
	defer c.CloseNow()
	ctx := c.CloseRead(context.Background())

	go func() {
		time.Sleep(time.Second)
		sb.msgs <- []byte("hello")

	}()

	for {
		select {
		case msg := <-sb.msgs:
			fmt.Printf("writed")
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (cs *ChatServer) addSubscriber(sb *Subscriber) {
	cs.subscriberMu.Lock()
	cs.subscribers[sb] = struct{}{}
	cs.subscriberMu.Unlock()

}

func (cs *ChatServer) deleteSubscriber(sb *Subscriber) {
	cs.subscriberMu.Lock()
	delete(cs.subscribers, sb)
	cs.subscriberMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
