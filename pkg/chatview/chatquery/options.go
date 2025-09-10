package chatquery

import "flag"

type Options struct {
	Port        string
	Host        string
	DatabaseURL string
}

func ConfigDefaultOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Port, "p", "8080", "Port to run the server on")
	fs.StringVar(&opts.Host, "H", "localhost", "Host to run the server on")
	fs.StringVar(&opts.DatabaseURL, "db-url", "redis://localhost:6379", "Database connection URL (default: redis://localhost:6379)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return &opts, nil
}
