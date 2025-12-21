package avatar

import (
	"errors"
	"flag"
	"os"
)

type Options struct {
	NatsUrl string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Region    string
	S3Bucket    string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL (env: NATS_URL)")

	fs.StringVar(&opts.S3Endpoint, "s3-endpoint", "", "S3/MinIO endpoint (env: S3_ENDPOINT)")
	fs.StringVar(&opts.S3AccessKey, "s3-access-key", "", "S3 access key (env: S3_ACCESS_KEY)")
	fs.StringVar(&opts.S3SecretKey, "s3-secret-key", "", "S3 secret key (env: S3_SECRET_KEY)")
	fs.StringVar(&opts.S3Region, "s3-region", "us-east-1", "S3 region (env: S3_REGION)")
	fs.StringVar(&opts.S3Bucket, "s3-bucket", "avatars", "S3 bucket for generated avatars (env: S3_BUCKET)")

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

	if opts.S3AccessKey == "" {
		if envVal := os.Getenv("S3_ACCESS_KEY"); envVal != "" {
			opts.S3AccessKey = envVal
		} else {
			opts.S3AccessKey = "minioadmin"
		}
	}

	if opts.S3SecretKey == "" {
		if envVal := os.Getenv("S3_SECRET_KEY"); envVal != "" {
			opts.S3SecretKey = envVal
		} else {
			opts.S3SecretKey = "minioadmin"
		}
	}

	if opts.S3Region == "" {
		if envVal := os.Getenv("S3_REGION"); envVal != "" {
			opts.S3Region = envVal
		} else {
			opts.S3Region = "us-east-1"
		}
	}

	if opts.S3Bucket == "" {
		if envVal := os.Getenv("S3_BUCKET"); envVal != "" {
			opts.S3Bucket = envVal
		} else {
			opts.S3Bucket = "avatars"
		}
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	return &opts, nil
}

func (o *Options) Validate() error {
	if o.NatsUrl == "" {
		return errors.New("nats url cannot be empty")
	}
	if o.S3Endpoint == "" {
		return errors.New("s3 endpoint cannot be empty")
	}
	if o.S3AccessKey == "" {
		return errors.New("s3 access key cannot be empty")
	}
	if o.S3SecretKey == "" {
		return errors.New("s3 secret key cannot be empty")
	}
	if o.S3Region == "" {
		return errors.New("s3 region cannot be empty")
	}
	if o.S3Bucket == "" {
		return errors.New("s3 bucket cannot be empty")
	}
	return nil
}
