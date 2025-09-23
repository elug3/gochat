package ws

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	Port string
	Host string

	// Authentication server url for token validation
	AuthServerUrl string
	// User info server url for fetching user chat info
	ContactsServerUrl string

	NatsUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host to bind the WebSocket server to")
	fs.StringVar(&opts.Port, "port", "12345", "Port to bind the WebSocket server to")

	fs.StringVar(&opts.AuthServerUrl, "auth-server", "", "Authentication server url")
	fs.StringVar(&opts.ContactsServerUrl, "contacts-server", "", "User info server url")
	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server url")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.AuthServerUrl == "" {
		if envVar := os.Getenv(("AUTH_SERVER")); envVar != "" {
			opts.AuthServerUrl = envVar
		} else {
			return nil, fmt.Errorf("Please set $AUTH_SERVER environment variable")
		}
	}
	if opts.ContactsServerUrl == "" {
		if envVar := os.Getenv(("CONTACTS_SERVER")); envVar != "" {
			opts.ContactsServerUrl = envVar
		} else {
			return nil, fmt.Errorf("Please set $CONTACTS_SERVER environment variable")
		}
	}
	if opts.NatsUrl == "" {
		if envVar := os.Getenv(("NATS_URL")); envVar != "" {
			opts.NatsUrl = envVar
		} else {
			return nil, fmt.Errorf("Please set $NATS_URL environment variable")
		}
	}

	return &opts, nil
}
