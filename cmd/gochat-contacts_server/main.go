package main

import (
	"flag"
	"fmt"
	"os"

	contacts "github.com/elug3/gochat/internal/services/contcts_service"
)

var usageStr = `
Usage: gochat-contacts_server [option]
	Server Options:
	-H --host <host>        Server host (default: "0.0.0.0")
	-P --port <port>        Server port (default: "8080")
	-d --data <data_dir>    Data directory (default: "./data")
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
		os.Exit(1)
	}
	srv, err := contacts.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("failed to create server: %v\n", err)
		os.Exit(1)
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
