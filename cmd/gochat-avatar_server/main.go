package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/elug3/gochat/pkg/avatar"
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	opts, err := avatar.ConfigureOptions(fs, os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	srv, err := avatar.NewServer(opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()
}
