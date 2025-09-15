package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/ws"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	fs := flag.NewFlagSet("ws-server", flag.ExitOnError)

	opts, err := ws.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("Error configuring options: %v\n", err)
		os.Exit(1)
	}

	hub := ws.NewHub()
	srv, err := ws.NewWsServer(hub, opts)
	if err != nil {
		fmt.Printf("Error creating WebSocket server: %v\n", err)
		os.Exit(1)
	}

	go func() {
		hub.Run(ctx)
	}()

	log.Info().Msgf("Starting WebSocket server on %s", opts.Port)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Error starting WebSocket server: %v\n", err)
		os.Exit(1)
	}
}
