package contacts

import (
	"flag"
	"os"
)

type Options struct {
	Host string
	Port string

	// directory where to save the database file
	SaveDir string
	// If true, the database will not be saved to disk
	NoSave bool
	// If true, the event will not be published
	NoEvent bool
	// If true, the user icons will not be generated
	NoIcons bool

	NatsUrl string

	S3Endpoint string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host address (default: 0.0.0.0)")
	fs.StringVar(&opts.Port, "p", "8080", "Port number (default: 8080)")
	fs.StringVar(&opts.Port, "port", "8080", "Port number (default: 8080)")

	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory where to save the database file")
	fs.BoolVar(&opts.NoSave, "no-save", false, "If true, the database will not be saved to disk")
	fs.BoolVar(&opts.NoEvent, "no-event", false, "If true, the event will not be published to NATS")
	fs.BoolVar(&opts.NoIcons, "no-icons", false, "If true, the user icons will not be generated")

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

	if opts.S3Endpoint == "" {
		if envVal := os.Getenv("S3_ENDPOINT"); envVal != "" {
			opts.S3Endpoint = envVal
		} else {
			opts.S3Endpoint = "http://localhost:9000"
		}
	}

	return &opts, nil
}
