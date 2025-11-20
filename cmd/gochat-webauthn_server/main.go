package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/webauthn"
)

func main() {
	fs := flag.NewFlagSet("webauthn", flag.ExitOnError)
	opts, err := webauthn.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("Error configuring options: %v\n", err)
		os.Exit(1)
	}
	srv, err := webauthn.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("Error creating WebAuthn server: %v\n", err)
		os.Exit(1)
	}

	if err = srv.ListenAndServe(); err != nil {
		fmt.Printf("Error starting WebAuthn server: %v\n", err)
		os.Exit(1)
	}
}
