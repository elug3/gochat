package contacts

import (
	"flag"
	"os"
)

type Options struct {
	Host string
	Post string

	// directory where to save the database file
	SaveDir string
	NoSave  bool

	NatsUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Host address (default: 0.0.0.0)")
	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host address (default: 0.0.0.0)")
	fs.StringVar(&opts.Post, "p", "8080", "Post number (default: 8080)")
	fs.StringVar(&opts.Post, "post", "8080", "Post number (default: 8080)")

	fs.StringVar(&opts.SaveDir, "d", "./", "Directory where to save the database file")
	fs.BoolVar(&opts.NoSave, "no-save", false, "Directory where to save the database file")

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.NatsUrl == "" {
		if envVal := os.Getenv("NATS_URL"); envVal != "" {
			opts.NatsUrl = envVal
		} else {
			opts.NatsUrl = "nats://localhost:4222"
		}
	}

	return &opts, nil
}
