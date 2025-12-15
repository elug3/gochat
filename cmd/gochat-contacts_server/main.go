package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	fs := flag.NewFlagSet("gochat-contacts_server", flag.ExitOnError)
	fs.Usage = printUsage
	opts, err := contacts.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("failed to configure options: %v\n", err)
		os.Exit(1)
	}
	srv, service, err := contacts.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}
	_ = service // service is available for future use
	
	eventListener, err := contacts.NewEventListener(opts)
	if err != nil {
		fmt.Printf("failed to create event listener: %v\n", err)
		os.Exit(1)
	}

	// Set up graceful shutdown
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := eventListener.Run(shutdownCtx)
		if err != nil {
			fmt.Printf("event listener error: %v\n", err)
			cancel()
		}
	}()

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		fmt.Printf("received signal: %v, shutting down gracefully...\n", sig)
		cancel()
		
		// Shutdown server with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30)
		defer cancel()
		if err := contacts.ShutdownServer(shutdownCtx, srv); err != nil {
			fmt.Printf("server shutdown error: %v\n", err)
		}
	case err := <-serverErrChan:
		fmt.Printf("server error: %v\n", err)
		cancel()
		os.Exit(1)
	}
}
