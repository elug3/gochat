package chatcmd

import (
	"flag"
)

type Options struct {
	DatabaseUrl string
	NatsUrl     string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	opts := &Options{}

	fs.StringVar(&opts.NatsUrl, "nats", "nats://localhost:4222", "NATS server URL (default: nats://localhost:4222)")
	fs.StringVar(&opts.DatabaseUrl, "db", "localhost:6379", "Database server URL (default: localhost:6379)")
	fs.Parse(args)

	return opts, nil
}
