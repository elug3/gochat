package message

import (
	"flag"
	"testing"
)

func TestConfigureOptions_RequiresContactsServer(t *testing.T) {
	t.Parallel()

	t.Setenv("CONTACTS_SERVER", "")
	t.Setenv("NATS_URL", "")

	if _, err := ConfigureOptions(flag.NewFlagSet("message", flag.ContinueOnError), nil); err == nil {
		t.Fatalf("expected error when contacts server is not configured")
	}
}

func TestConfigureOptions_DefaultsAndEnv(t *testing.T) {
	t.Parallel()

	t.Setenv("CONTACTS_SERVER", "http://contacts:8080")
	t.Setenv("NATS_URL", "nats://env-nats:4222")

	opts, err := ConfigureOptions(flag.NewFlagSet("message", flag.ContinueOnError), nil)
	if err != nil {
		t.Fatalf("ConfigureOptions: %v", err)
	}

	if opts.Port != "8080" {
		t.Fatalf("unexpected Port: %q", opts.Port)
	}
	if opts.Host != "0.0.0.0" {
		t.Fatalf("unexpected Host: %q", opts.Host)
	}
	if opts.SaveDir != "./" {
		t.Fatalf("unexpected SaveDir: %q", opts.SaveDir)
	}
	if opts.NoSave {
		t.Fatalf("expected NoSave to be false by default")
	}
	if opts.NatsUrl != "nats://env-nats:4222" {
		t.Fatalf("unexpected NatsUrl: %q", opts.NatsUrl)
	}
	if opts.ContactsServerUrl != "http://contacts:8080" {
		t.Fatalf("unexpected ContactsServerUrl: %q", opts.ContactsServerUrl)
	}
}

func TestConfigureOptions_UsesFlagOverrides(t *testing.T) {
	t.Parallel()

	t.Setenv("CONTACTS_SERVER", "http://contacts:8080")
	t.Setenv("NATS_URL", "")

	fs := flag.NewFlagSet("message", flag.ContinueOnError)
	args := []string{"-contacts-server", "http://flag-contacts:9090", "-nats-url", "nats://flag-nats:4222", "-no-save", "-port", "9091", "-host", "127.0.0.1", "-save-dir", "/tmp/messages"}

	opts, err := ConfigureOptions(fs, args)
	if err != nil {
		t.Fatalf("ConfigureOptions with flags: %v", err)
	}

	if opts.Port != "9091" {
		t.Fatalf("unexpected Port: %q", opts.Port)
	}
	if opts.Host != "127.0.0.1" {
		t.Fatalf("unexpected Host: %q", opts.Host)
	}
	if opts.SaveDir != "/tmp/messages" {
		t.Fatalf("unexpected SaveDir: %q", opts.SaveDir)
	}
	if !opts.NoSave {
		t.Fatalf("expected NoSave to be true when flag set")
	}
	if opts.NatsUrl != "nats://flag-nats:4222" {
		t.Fatalf("unexpected NatsUrl: %q", opts.NatsUrl)
	}
	if opts.ContactsServerUrl != "http://flag-contacts:9090" {
		t.Fatalf("unexpected ContactsServerUrl: %q", opts.ContactsServerUrl)
	}
}
