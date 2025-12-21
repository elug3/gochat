package avatar

import (
	"flag"
	"testing"
)

func TestOptionsValidate(t *testing.T) {
	t.Parallel()

	_, err := ConfigureOptions(flag.NewFlagSet("avatar", flag.ContinueOnError), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = (&Options{}).Validate()
	if err == nil {
		t.Fatalf("expected error for empty options")
	}
}

func TestConfigureOptions_Defaults(t *testing.T) {
	t.Setenv("NATS_URL", "")
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_ACCESS_KEY", "")
	t.Setenv("S3_SECRET_KEY", "")
	t.Setenv("S3_REGION", "")
	t.Setenv("S3_BUCKET", "")

	opts, err := ConfigureOptions(flag.NewFlagSet("avatar", flag.ContinueOnError), nil)
	if err != nil {
		t.Fatalf("ConfigureOptions: %v", err)
	}

	if opts.NatsUrl != "nats://localhost:4222" {
		t.Fatalf("unexpected NatsUrl: %q", opts.NatsUrl)
	}
	if opts.S3Endpoint != "http://localhost:9000" {
		t.Fatalf("unexpected S3Endpoint: %q", opts.S3Endpoint)
	}
	if opts.S3AccessKey != "minioadmin" {
		t.Fatalf("unexpected S3AccessKey: %q", opts.S3AccessKey)
	}
	if opts.S3SecretKey != "minioadmin" {
		t.Fatalf("unexpected S3SecretKey: %q", opts.S3SecretKey)
	}
	if opts.S3Region != "us-east-1" {
		t.Fatalf("unexpected S3Region: %q", opts.S3Region)
	}
	if opts.S3Bucket != "avatars" {
		t.Fatalf("unexpected S3Bucket: %q", opts.S3Bucket)
	}
}
