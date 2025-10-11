package presence

import (
	"flag"
	"fmt"
	"os"
)

type HttpOptions struct {
	Host      string
	Port      string
	RedisAddr string
}

type EventOptions struct {
	NatsUrl string
}

func ConfigureHttpOptions(fs *flag.FlagSet, args []string) (*HttpOptions, error) {
	var opts HttpOptions

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "HTTP server host")
	fs.StringVar(&opts.Port, "port", "8080", "HTTP server port")
	fs.StringVar(&opts.RedisAddr, "redis-url", "", "redis address for presence store")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if opts.RedisAddr == "" {
		if redisAddr, exists := os.LookupEnv("REDIS_URL"); exists {
			opts.RedisAddr = redisAddr
		} else {
			return nil, fmt.Errorf("redis address must be provided via --redis-url flag or REDIS_URL environment variable")
		}
	}

	return &opts, nil
}

func ConfigureEventOptions(fs *flag.FlagSet, args []string) (*EventOptions, error) {
	var opts EventOptions

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if opts.NatsUrl == "" {
		if natsUrl, exists := os.LookupEnv("NATS_URL"); exists {
			opts.NatsUrl = natsUrl
		} else {
			return nil, fmt.Errorf("NATS URL must be provided via --nats-url flag or NATS_URL environment variable")
		}
	}

	return &opts, nil
}
