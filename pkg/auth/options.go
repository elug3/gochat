package auth

import (
	"flag"
	"os"

	"github.com/rs/zerolog/log"
)

type Options struct {
	Host string
	Port string

	// Path to RSA private key in PEM format
	KeyPath   string
	UseTmpKey bool

	SaveDir  string
	InMemory bool

	// debug disables recovery middleware
	debug bool

	NatsUrl  string
	NoEvents bool

	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options
	var (
		showHelp    bool
		showVersion bool
	)

	fs.BoolVar(&showHelp, "h", false, "Show help")
	fs.BoolVar(&showVersion, "v", false, "Show version")

	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Server host")
	fs.StringVar(&opts.Port, "p", "8080", "Server port")

	fs.StringVar(&opts.KeyPath, "secret", "", "Path to RSA private key in PEM format")
	fs.BoolVar(&opts.UseTmpKey, "tmpkey", false, "Use temporary key")

	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory to save the database (default: ./) ")
	fs.BoolVar(&opts.InMemory, "in-memory", false, "Use in-memory database")

	fs.StringVar(&opts.NatsUrl, "nats-url", "", "NATS server URL")
	fs.BoolVar(&opts.NoEvents, "no-events", false, "Disable event publishing")

	fs.BoolVar(&opts.debug, "debug", false, "Enable debug mode")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showVersion {
		// Print version information
	}

	if showHelp {
		fs.Usage()
		return nil, nil
	}

	if opts.NatsUrl == "" {
		if envVar, ok := os.LookupEnv("NATS_URL"); ok {
			opts.NatsUrl = envVar
		} else {
			opts.NatsUrl = "nats://localhost:4222"
			log.Warn().Msgf("no NATS URL provided, defaulting to %s", opts.NatsUrl)
		}
	}

	if opts.KeyPath == "" && !opts.UseTmpKey {
		log.Warn().Msg("no key path provided, using temporary key")
		opts.UseTmpKey = true
	} else {
		log.Info().Msg("using provided key path")
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

	return &opts, nil
}
