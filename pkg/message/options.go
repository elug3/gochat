package message

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	Port               string
	Host               string
	SaveDir            string
	NoSave             bool
	NatsUrl            string
	ContactsServiceUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Port, "port", "8080", "Port to run the server on")
	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host to run the server on")
	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory to save data")
	fs.BoolVar(&opts.NoSave, "no-save", false, "Run in no-save mode (data will not be persisted)")
	fs.StringVar(&opts.NatsUrl, "nats-url", "", "URL for the NATS server (default: nats://localhost:4222)")
	fs.StringVar(&opts.ContactsServiceUrl, "contacts-server", "", "Host for the contacts service")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if opts.NatsUrl == "" {
		if envVar := os.Getenv("NATS_URL"); envVar != "" {
			opts.NatsUrl = envVar
		} else {
			opts.NatsUrl = "nats://localhost:4222"
		}
	}

	if opts.ContactsServiceUrl == "" {
		if envVar := os.Getenv("CONTACTS_SERVER"); envVar != "" {
			opts.ContactsServiceUrl = envVar
		} else {
			return nil, fmt.Errorf("contacts-server option or CONTACTS_SERVER environment variable must be set")
		}
	}

	return &opts, nil
}
