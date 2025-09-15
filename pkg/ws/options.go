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
	AuthUrl string
	// User info server url for fetching user chat info
	ContactsUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host to bind the WebSocket server to")
	fs.StringVar(&opts.Port, "port", "12345", "Port to bind the WebSocket server to")

	fs.StringVar(&opts.AuthUrl, "auth-url", "", "Authentication server url")
	fs.StringVar(&opts.ContactsUrl, "contacts-url", "", "User info server url")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.AuthUrl == "" {
		if envVar := os.Getenv(("AUTH_URL")); envVar != "" {
			opts.AuthUrl = envVar
		} else {
			return nil, fmt.Errorf("Please set $AUTH_URL environment variable")
		}
	}
	if opts.ContactsUrl == "" {
		if envVar := os.Getenv(("CONTACTS_URL")); envVar != "" {
			opts.ContactsUrl = envVar
		} else {
			return nil, fmt.Errorf("Please set $CONTACTS_URL environment variable")
		}
	}

	return &opts, nil
}
