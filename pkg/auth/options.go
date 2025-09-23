package auth

import (
	"flag"

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

	if opts.KeyPath == "" && !opts.UseTmpKey {
		log.Warn().Msg("no key path provided, using temporary key")
		opts.UseTmpKey = true
	} else {
		log.Info().Msg("using provided key path")
	}

	return &opts, nil
}
