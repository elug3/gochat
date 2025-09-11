package chatcmd

import (
	"flag"
	"os"
)

type Options struct {
	DatabaseUrl string
	NatsUrl     string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := &Options{}

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL (default: nats://localhost:4222)")
	fs.StringVar(&opts.DatabaseUrl, "db", "localhost:6379", "Database server URL (default: localhost:6379)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if opts.NatsUrl == "" {
		if envVal := os.Getenv("NATS_URL"); envVal != "" {
			opts.NatsUrl = envVal
		} else {
			opts.NatsUrl = "nats://localhost:4222"
		}
	}

	return opts, nil
}
