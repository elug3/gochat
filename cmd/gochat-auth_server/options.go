package main

import (
	"flag"
	"os"

	"github.com/elug3/gochat/pkg/auth"
)

var usageStr = `
	Usage: gochat-auth [options]

	Server options:
		--host <host>			Bind address for the HTTP server (default: 0.0.0.0)
		--port <port>			Port for the HTTP server (default: 8080)

	JWT / keys:
		--secret <path>			Path to RSA private key (PEM). If omitted and
								--tmpkey is not set, a temporary key will be used.
		--tmpkey				Use a temporary in-memory JWT signing key

	Storage / database:
		--save-dir <dir>		Directory to save the database (default: ./)
		--in-memory				Use an in-memory database (data will not persist)

	Events / messaging:
		--nats-url <url>		NATS server URL (default: from $NATS_URL or nats://localhost:4222)
		--no-events				Disable event publishing to NATS

	Logging:
	    --log-level=LEVEL		set logging verbosity
								LEVEL is one of: debug, info, warn (default: info)
		--log-format=FORMAT		set log output format
								FORMAT IS one of: json, text (defult: json)
	General:
		--help					Show this help message
		--version				Show version information

	Environment variables:
		NATS_URL				Alternative to --nats-url
		WEBAUTHN_RP_DISPLAY_NAME  Relying Party display name (default: GoChat)
		WEBAUTHN_RPID			Relying Party id (default: localhost)
		WEBAUTHN_RP_ORIGINS		Comma-separated list of allowed origins (default: http://localhost:5173)
`

func ConfigureOptions(fs *flag.FlagSet, args []string) (auth.Options, error) {
	var opts auth.Options
	var (
		showHelp    bool
		showVersion bool
	)

	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Show version")

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Server host")
	fs.StringVar(&opts.Port, "port", "8080", "Server port")

	fs.StringVar(&opts.JWTKeyPath, "secret", "", "Path to RSA private key in PEM format")
	fs.BoolVar(&opts.UseTemporaryJWTKey, "tmpkey", false, "Use temporary key")

	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory to save the database (default: ./) ")
	fs.BoolVar(&opts.InMemory, "in-memory", false, "Use in-memory database")

	fs.StringVar(&opts.LogLevel, "log-level", "info", "set logging verbosity (default: info)")
	fs.StringVar(&opts.LogFormat, "log-format", "json", "set log output format (default: json)")

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL")
	fs.BoolVar(&opts.NoEvents, "no-events", false, "Disable event publishing")

	if err := fs.Parse(args); err != nil {
		return auth.Options{}, err
	}

	if showVersion {
		// Print version information
	}

	if showHelp {
		fs.Usage()
		return auth.Options{}, nil
	}

	if opts.NatsUrl == "" {
		if envVar, ok := os.LookupEnv("NATS_URL"); ok {
			opts.NatsUrl = envVar
		} else {
			opts.NatsUrl = "nats://localhost:4222"
		}
	}

	if opts.JWTKeyPath == "" && !opts.UseTemporaryJWTKey {
		opts.UseTemporaryJWTKey = true
	}

	if envVar, ok := os.LookupEnv("WEBAUTHN_RP_DISPLAY_NAME"); ok {
		opts.RPDisplayName = envVar
	} else {
		opts.RPDisplayName = "GoChat"
	}

	if envVar, ok := os.LookupEnv("WEBAUTHN_RPID"); ok {
		opts.RPID = envVar
	} else {
		opts.RPID = "localhost"
	}

	if envVar, ok := os.LookupEnv("WEBAUTHN_RP_ORIGINS"); ok {
		opts.RPOrigins = []string{envVar}
	} else {
		opts.RPOrigins = []string{"http://localhost:5173"}
	}

	return opts, nil
}
