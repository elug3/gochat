package main

import (
	"flag"
	"fmt"
	"os"

	usersrv "github.com/elug3/gochat/pkg/users"
)

var usageStr = `
Usage: gochat [options]

Server Options:
	-a <host>			Bind to host address (default: 0.0.0.0)
	-p <port>			Port to run the server on (default: 8080)

Command Options:
	-h			 		Show this help message
	-v			 		Show version
`

func usage() {
	fmt.Print(usageStr)
}

func main() {
	fs := flag.NewFlagSet("user-server", flag.ExitOnError)
	fs.Usage = usage

	opts, err := usersrv.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("cannot parse command line options: %v\n", err)
		os.Exit(2)
	}

	server, err := usersrv.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("cannot create user service server: %v\n", err)
		os.Exit(2)
	}

	if err = server.Run(); err != nil {
		fmt.Printf("failed to run user service server: %v\n", err)
		os.Exit(2)
	}
}
