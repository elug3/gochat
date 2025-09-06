package main

import (
	"flag"
)

var usageStr = `
Usage: gochat-auth [option]...

Server Options:
	-a <host>			Bind to host address (default: 0.0.0.0)
	-p <port>			Port to run the server on (default: 8080)

Command Options:
	-h			 		Show this help message
	-v			 		Show version
`

func printUsage() {
	print(usageStr)
}

func main() {
	fs := flag.NewFlagSet("gochat-auth_server", flag.ExitOnError)
	fs.Usage = printUsage

	// opts, err := authsrv.ConfiguireOptions(fs, os.Args[1:])
	// if err != nil {
	// 	fmt.Printf("cannot parse command line options: %v:\n", err)
	// 	printUsage()
	// 	os.Exit(1)
	// }

}
