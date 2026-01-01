package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	authsrv "github.com/elug3/gochat/pkg/auth"
)

func printUsage() {
	print(usageStr)
	os.Exit(1)
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fs := flag.NewFlagSet("gochat-auth_server", flag.ExitOnError)
	fs.Usage = printUsage

	opts, err := ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("cannot parse command line options: %v:\n", err)
		os.Exit(1)
	}

	srv, err := authsrv.NewServer(opts)
	if err != nil {
		fmt.Printf("cannot create auth server: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Run(ctx); err != nil {
		os.Exit(1)
	}
}
