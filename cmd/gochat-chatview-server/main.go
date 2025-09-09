package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/elug3/gochat/pkg/chatview/chatquery"
)

var usageStr = ``

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
	srv.ListenAndServe()
}
