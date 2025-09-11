package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/chatview/chatcmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var usageStr = `
Usage: gochat-chatview_server [option]

Server Options:
	--db url	redis Database URL (default: redis://localhost:6379)
	--nats url	NATS server URL (default: nats://localhost:4222)
`

func printUsage() {
	fmt.Printf("%s\n", usageStr)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fs := flag.NewFlagSet("gochat-chatview_server", flag.ExitOnError)
	fs.Usage = printUsage

	opts, err := chatcmd.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("error configuring options: %v\n", err)
		os.Exit(1)
	}

	srv, err := chatcmd.NewChatProjectionServer(opts)
	if err != nil {
		fmt.Printf("error creating chat projection server: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Run(context.Background()); err != nil {
		fmt.Printf("error running chat projection server: %v\n", err)
		os.Exit(1)
	}
}
