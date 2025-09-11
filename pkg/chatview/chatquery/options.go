package chatquery

import (
	"flag"
	"os"
)

type Options struct {
	Port        string
	Host        string
	DatabaseURL string
}

func ConfigDefaultOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Port, "p", "8080", "Port to run the server on")
	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Host to run the server on")
	fs.StringVar(&opts.DatabaseURL, "db-url", "", "Database connection URL (default: localhost:6379)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if opts.DatabaseURL == "" {
		if envVar := os.Getenv("DB_URL"); envVar != "" {
			opts.DatabaseURL = envVar
		} else {
			opts.DatabaseURL = "localhost:6379"
		}
	}

	return &opts, nil
}
