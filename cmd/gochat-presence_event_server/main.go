package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/presence"
)

func main() {
	ctx := context.Background()
	fs := flag.NewFlagSet("gochat-presence_event_server", flag.ExitOnError)

	opts, err := presence.ConfigureEventOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("error configuring options: %v\n", err)
		os.Exit(1)
	}

	l, err := presence.NewEventListener(opts)
	if err != nil {
		fmt.Printf("error creating event listener: %v\n", err)
		os.Exit(1)
	}
	if err = l.Listen(ctx); err != nil {
		fmt.Printf("error listening for events: %v\n", err)
		os.Exit(1)
	}
}
