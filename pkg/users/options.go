package users

import (
	"flag"
	"fmt"
)

type Options struct {
	Host string
	Port string
}

func printVersion() {
	fmt.Print("version")
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options
	var (
		showVersion bool
		showHelp    bool
	)
	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Show version")
	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Server host")
	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Server host")
	fs.StringVar(&opts.Port, "p", "8080", "Server port")
	fs.StringVar(&opts.Port, "port", "8080", "Server port")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showVersion {
		printVersion()
		return nil, nil
	}
	if showHelp {
		fs.Usage()
		return nil, nil
	}

	return &opts, nil
}
