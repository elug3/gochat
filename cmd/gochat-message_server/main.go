package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/message"
)

var usageStr = `
Usage: gochat-message_server [options]
	Server Options:
		-port <port>               Port to run the server on (default: 8080)
		-host <host>               Host to run the server on (default: 0.0.0.0)
		-save-dir <dir>           Directory to save data (default: ./ )
		-no-save                  Run in no-save mode (data will not be persisted) (default: false)
		-nats-url <url>           URL for the NATS server (default: nats://localhost:4222)
		-contacts-server <host>       Host for the contacts service (default: localhost:8080)
`

func printUsage() {
	fmt.Println(usageStr)
}

func main() {
	ctx := context.Background()
	fs := flag.NewFlagSet("gochat-message_server", flag.ExitOnError)
	fs.Usage = printUsage
	opts, err := message.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("Error configuring options: %v\n", err)
		os.Exit(1)
	}

	srv, err := message.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		os.Exit(1)
	}
	eventSrv, err := message.NewEventServer(opts)
	if err != nil {
		fmt.Printf("Error creating event server: %v\n", err)
	}

	go func() {
		fmt.Printf("event listener started\n")
		if err := eventSrv.Run(ctx); err != nil {
			fmt.Printf("event server: %v\n", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
