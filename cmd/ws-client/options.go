package main

import "flag"

type Options struct {
	WsUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.WsUrl, "ws-url", "ws://localhost:8080/ws", "WebSocket URL of the API")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return &opts, nil
}
