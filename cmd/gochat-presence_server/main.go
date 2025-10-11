package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/presence"
)

func main() {

	fs := flag.NewFlagSet("gochat-presence_server", flag.ExitOnError)
	opts, err := presence.ConfigureHttpOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("Error configuring options: %v\n", err)
		return
	}

	srv, err := presence.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		return
	}

	if err = srv.ListenAndServe(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
		return
	}
}
