package ws_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type BenchMsg struct {
	Content string
	WG      *sync.WaitGroup
}

type Client struct {
	Id         int
	WriteDelay time.Duration
}

type User struct {
	Id      int32
	Clients map[int]*Client

	bcastCh chan *BenchMsg
}

type Chat struct {
	Users map[int32]*User
}

func newChat() *Chat {
	return &Chat{
		Users: make(map[int32]*User),
	}
}

func (c *Client) Write(msg *BenchMsg) {
	if c.WriteDelay > 0 {
		time.Sleep(c.WriteDelay)
	}
	msg.WG.Done()
}

const (
	Users           = 100
	SessionsPerUser = 10
	PayloadSize     = 128
)

func (u *User) StartBroadcaster(ctx context.Context) {
	u.bcastCh = make(chan *BenchMsg, 256)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-u.bcastCh:
				for _, c := range u.Clients {
					c.Write(msg)
				}
			}
		}
	}()
}

func (u *User) StopBroadcaster() {
	if u.bcastCh != nil {
		close(u.bcastCh)
		u.bcastCh = nil
	}
}

// --- Baseline broadcast: hub iterates users->clients and writes synchronously ---
func baselineBroadcast(chat *Chat, msg *BenchMsg) {
	for _, user := range chat.Users {
		for _, client := range user.Clients {
			client.Write(msg) // synchronous write
		}
	}
}

// --- Per-user broadcaster publish: hub sends into each user's channel ---
func perUserBroadcast(chat *Chat, msg *BenchMsg) {
	for _, user := range chat.Users {
		// if a broadcaster doesn't exist, fallback to direct write
		if user.bcastCh != nil {
			user.bcastCh <- msg
		} else {
			for _, c := range user.Clients {
				c.Write(msg)
			}
		}
	}
}

func printGoroutines(prefix string) {
	fmt.Printf("%s goroutines=%d\n", prefix, runtime.NumGoroutine())
}

func makeUsers(U, S int, writeDelay time.Duration) (map[int32]*User, int) {
	users := make(map[int32]*User, U)
	clientId := 0

	for i := 0; i < U; i++ {
		u := &User{
			Id:      int32(i + 1),
			Clients: make(map[int]*Client, S),
		}
		for j := 0; j < S; j++ {
			clientId++
			u.Clients[clientId] = &Client{Id: clientId, WriteDelay: writeDelay}
		}
		users[u.Id] = u
	}
	total := U * S

	return users, total
}

func Benchmark_Hub(b *testing.B) {
	var writeDelay time.Duration = 0 * time.Millisecond

	users, totalClients := makeUsers(Users, SessionsPerUser, writeDelay)
	chat := newChat()
	chat.Users = users

	payload := strings.Repeat("x", PayloadSize)

	b.Run("baseline-broadcast", func(b *testing.B) {

		printGoroutines("before baseline")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			wg := &sync.WaitGroup{}
			wg.Add(totalClients)
			msg := &BenchMsg{Content: payload, WG: wg}
			baselineBroadcast(chat, msg)
			wg.Wait()
		}
		b.StopTimer()
		printGoroutines("after baseline")
	})
}
