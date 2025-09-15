package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elug3/gochat/api/httpclient"
	"github.com/elug3/gochat/cmd/ws-client/console"
)

var usageStr = `
Usage: ws-client [options]

Options:
	-h, --help          Show this help message and exit
	-t, --token TOKEN   API token (or set the GOCHAT_API_TOKEN environment variable)
	-u, --url URL       Base URL of the API (default: http://localhost:8080)
	--ws, URL 	  WebSocket URL of the API (default: ws://localhost:8080/ws)
`

func printUsage() {
	fmt.Println(usageStr)
}

var api *httpclient.Api

func main() {
	fs := flag.NewFlagSet("ws-client", flag.ExitOnError)
	fs.Usage = printUsage

	opts, err := httpclient.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("Error parsing options: %v\n", err)
		return
	}

	api = httpclient.NewApi(opts...)

	ctx := context.Background()
	if ok, err := api.IsAuthenticated(ctx); !ok {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Not authenticated. Please check your api token.")
		return
	}

	list, _ := console.NewList(nil, loadGroupList)

	console.Clear()
	if _, err := tea.NewProgram(list).Run(); err != nil {
		fmt.Println("Error running program:", err)
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
