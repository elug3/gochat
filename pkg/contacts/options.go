package contacts

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultIdleTimeout  = 120 * time.Second
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

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Region    string
	IconBaseURL string

	// HTTP server timeouts
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Host address (default: 0.0.0.0)")
	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Host address (default: 0.0.0.0)")
	fs.StringVar(&opts.Port, "p", "8080", "Port number (default: 8080)")
	fs.StringVar(&opts.Port, "port", "8080", "Port number (default: 8080)")

	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory where to save the database file")
	fs.BoolVar(&opts.NoSave, "no-save", false, "If true, the database will not be saved to disk")
	fs.BoolVar(&opts.NoEvent, "no-event", false, "If true, the event will not be published to NATS")
	fs.BoolVar(&opts.NoIcons, "no-icons", false, "If true, the user icons will not be generated")

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL")
	fs.StringVar(&opts.S3Endpoint, "s3-endpoint", "", "S3/MinIO endpoint for profile icons")
	fs.StringVar(&opts.S3AccessKey, "s3-access-key", "", "S3 access key (env: S3_ACCESS_KEY)")
	fs.StringVar(&opts.S3SecretKey, "s3-secret-key", "", "S3 secret key (env: S3_SECRET_KEY)")
	fs.StringVar(&opts.S3Region, "s3-region", "us-east-1", "S3 region (env: S3_REGION)")
	fs.StringVar(&opts.IconBaseURL, "icon-base-url", "", "Base URL for icon access (env: ICON_BASE_URL)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Set defaults from environment variables
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

	if opts.S3AccessKey == "" {
		opts.S3AccessKey = os.Getenv("S3_ACCESS_KEY")
		if opts.S3AccessKey == "" {
			opts.S3AccessKey = "minioadmin" // fallback for development
		}
	}

	if opts.S3SecretKey == "" {
		opts.S3SecretKey = os.Getenv("S3_SECRET_KEY")
		if opts.S3SecretKey == "" {
			opts.S3SecretKey = "minioadmin" // fallback for development
		}
	}

	if opts.S3Region == "" {
		if envVal := os.Getenv("S3_REGION"); envVal != "" {
			opts.S3Region = envVal
		} else {
			opts.S3Region = "us-east-1"
		}
	}

	if opts.IconBaseURL == "" {
		if envVal := os.Getenv("ICON_BASE_URL"); envVal != "" {
			opts.IconBaseURL = envVal
		} else {
			opts.IconBaseURL = opts.S3Endpoint // fallback to S3 endpoint
		}
	}

	// Set default timeouts
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = DefaultReadTimeout
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = DefaultWriteTimeout
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = DefaultIdleTimeout
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &opts, nil
}

// Validate checks that all required options are set and valid
func (o *Options) Validate() error {
	if o.Host == "" {
		return errors.New("host cannot be empty")
	}

	if o.Port == "" {
		return errors.New("port cannot be empty")
	}

	portNum, err := strconv.Atoi(o.Port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", portNum)
	}

	return nil
}
