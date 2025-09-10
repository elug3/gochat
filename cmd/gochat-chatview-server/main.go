package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/chatview/chatquery"
)

var usageStr = `
Usage:
  -p, --port string       Port to run the server on (default "8080")
  -H, --host string       Host to run the server on (default "localhost")
  -db-url string          Database connection URL (default: redis://localhost:6379)
  -v, --version           Show version information
`

func printUsage() {
	println(usageStr)
}

func main() {
	fs := flag.NewFlagSet("gochat-chatview_server", flag.ExitOnError)
	fs.Usage = printUsage

	opts, err := chatquery.ConfigDefaultOptions(fs, os.Args[1:])
	if err != nil {
		fmt.Printf("cannot parse options: %v\n", err)
		os.Exit(1)
	}

	srv, err := chatquery.NewHttpServer(opts)
	if err != nil {
		fmt.Printf("cannot create server: %v\n", err)
		os.Exit(1)
	}
	if err = srv.ListenAndServe(); err != nil {
		fmt.Printf("cannot start server: %v\n", err)
		os.Exit(1)
	}
}
