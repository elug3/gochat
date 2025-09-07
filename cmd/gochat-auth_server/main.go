package main

import (
	"flag"
	"fmt"
	"os"

	authsrv "github.com/elug3/gochat/pkg/auth"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var usageStr = `
Usage: gochat-auth [option]...

Server Options:
	-H <host>			Bind to host address (default: 0.0.0.0)
	-p <port>			Port to run the server on (default: 8080)

Command Options:
	-h			 		Show this help message
	-v			 		Show version
`

func printUsage() {
	print(usageStr)
	os.Exit(1)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fs := flag.NewFlagSet("gochat-auth_server", flag.ExitOnError)
	fs.Usage = printUsage

	opts, err := authsrv.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("cannot parse command line options: %v:\n", err)
		os.Exit(1)
	}

	s, err := authsrv.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("cannot create auth server: %v\n", err)
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		fmt.Printf("auth server error: %v\n", err)
		os.Exit(1)
	}
}
