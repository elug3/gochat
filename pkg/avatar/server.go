package avatar

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elug3/gochat/shared/events"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	eventSub *events.Subscriber
	logger   *zerolog.Logger
	handler  *Handler
}

func NewServer(opts *Options) (*Server, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opts.S3Endpoint == "" {
		return nil, fmt.Errorf("s3 endpoint cannot be empty")
	}
	if opts.S3Bucket == "" {
		return nil, fmt.Errorf("s3 bucket cannot be empty")
	}
	if _, err := url.ParseRequestURI(opts.S3Endpoint); err != nil {
		return nil, fmt.Errorf("invalid s3 endpoint %q: %w", opts.S3Endpoint, err)
	}

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	nc, err := nats.Connect(opts.NatsUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats at %q: %w", opts.NatsUrl, err)
	}

	sub, err := events.NewSubscriber(nc, &logger, events.APP_STREAM, &nats.ConsumerConfig{
		Durable:     events.DurableAvatar,
		AckPolicy:   nats.AckExplicitPolicy,
		Description: "avatar service consumer",
		FilterSubjects: []string{
			events.SubjectProfileCreated,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(opts.S3Region),
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: opts.S3Endpoint, HostnameImmutable: true}, nil
			}),
		),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.S3AccessKey, opts.S3SecretKey, ""),
		),
	)
	if err != nil {
		sub.Close()
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // required for MinIO in most setups
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(opts.S3Bucket)}); err != nil {
		sub.Close()
		return nil, fmt.Errorf("failed to access bucket %q at %s: %w", opts.S3Bucket, opts.S3Endpoint, err)
	}

	handler, err := NewHandler(s3Client, opts.S3Bucket)
	if err != nil {
		sub.Close()
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	srv := &Server{
		eventSub: sub,
		handler:  handler,
	}
	return srv, nil
}

func (srv *Server) Run(ctx context.Context) error {
	err := srv.eventSub.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run event subscriber: %w", err)
	}
	return err
}
