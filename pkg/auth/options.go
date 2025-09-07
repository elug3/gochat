package auth

import "flag"

type Options struct {
	Host string
	Port string

	// Path to RSA private key in PEM format
	KeyPath string
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

	fs.StringVar(&opts.KeyPath, "k", "jwt.key", "Path to RSA private key in PEM format (default './jwt.key')")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showVersion {
	}

	if showHelp {
		fs.Usage()
		return nil, nil
	}

	return &opts, nil
}
