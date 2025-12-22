package main

import (
	"flag"
	"os"
)

type Options struct {
	WsUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.WsUrl, "ws-url", "", "WebSocket URL of the API (env: WS_URL)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if opts.WsUrl == "" {
		if envVal := os.Getenv("WS_URL"); envVal != "" {
			opts.WsUrl = envVal
		} else {
			opts.WsUrl = "ws://localhost:8080/ws"
		}
	}

	return &opts, nil
}
