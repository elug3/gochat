package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/contacts"
)

var usageStr = `
Usage: gochat-contacts_server [option]
	Server Options:
	-H --host <host>        Server host (default: "0.0.0.0")
	-P --port <port>        Server port (default: "8080")
	-d --data <data_dir>    Data directory (default: "./ ")
	--help                  Show this message

	-nats <nats_url>        NATS server URL (default: "nats://localhost:4222")
`

func printUsage() {
	fmt.Printf("%s\n", usageStr)
}

func main() {
	ctx := context.Background()
	fs := flag.NewFlagSet("gochat-contacts_server", flag.ExitOnError)
	fs.Usage = printUsage
	opts, err := contacts.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
	srv, err := contacts.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}
	eventListener, err := contacts.NewEventListener(opts)
	if err != nil {
		fmt.Printf("failed to create event listener: %v\n", err)
		os.Exit(1)
	}

	go func() {
		err := eventListener.Run(ctx)
		if err != nil {
			fmt.Printf("event listener error: %v\n", err)
			os.Exit(1)
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
