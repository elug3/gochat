package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
	"github.com/elug3/gochat/api/httpclient"
	"github.com/elug3/gochat/cmd/ws-client/console"
)

var usageStr = `
Usage: ws-client [options]

Options:
	-h, --help          Show this help message and exit
	-t, --token TOKEN   API token (or set the GOCHAT_API_TOKEN environment variable)
	-u, --url URL       Base URL of the API (default: http://localhost:8080)
	-ws-url, URL       WebSocket URL of the API (default: ws://localhost:8080/ws)
`

func printUsage() {
	fmt.Println(usageStr)
}

var api *httpclient.Api

type Message struct {
	ChatId  int    `json:"chat_id"`
	Content string `json:"content"`
}

var wsUrl = os.Getenv("GOCHAT_WS_URL")

func main() {
	ctx := context.Background()

	conn, resp, err := websocket.Dial(ctx, wsUrl, &websocket.DialOptions{})
	if err != nil {
		fmt.Printf("Failed to connect to WebSocket: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("cannot read body: %v\n", err)
	}
	if resp.StatusCode != 101 {
		fmt.Printf("Unexpected status code: %d: %s\n", resp.StatusCode, body)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	go func() {
		for {
			msg := Message{ChatId: 1, Content: "Hello, WebSocket!"}
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Error marshaling message: %v\n", err)
				return
			}
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				fmt.Printf("Error writing message: %v\n", err)
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			return
		}
		fmt.Printf("Received: %s\n", msg)
	}
}

func loadGroupList(ctx context.Context) ([]console.ListItem, error) {
	group, err := api.Contacts.Groups.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load group list: %w", err)
	}
	items := make([]console.ListItem, len(group))
	for i, g := range group {
		items[i] = console.ListItem{
			Name:   g.Name,
			OpenFn: chatOpener(g.Id),
		}
	}
	return items, nil
}
func chatOpener(chatId int) console.OpenFunc {
	return func(prev tea.Model) (tea.Model, tea.Cmd) {
		loader := func(ctx context.Context) ([]console.Content, error) {
			messages, err := api.Messages.List(ctx, chatId)
			if err != nil {
				return nil, fmt.Errorf("failed to load messages: %w", err)
			}
			contents := make([]console.Content, len(messages))
			for i, msg := range messages {
				contents[i] = console.Content{
					Id:     msg.Id,
					Sender: msg.Sender,
					Time:   msg.SentAt,
					Text:   msg.Content,
				}
			}
			return contents, nil
		}
		return console.NewViewerWithLoader(prev, loader)
	}
}
