package auth

import "flag"

type Options struct {
	Host string
	Port string
}

func ConfigureOptions(fs flag.FlagSet) (*Options, error) {
	var opts Options
	var (
		showHelp    bool
		showVersion bool
	)

	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Show version")

	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Server host")
	fs.StringVar(&opts.Port, "port", "8080", "Server port")

	if err := fs.Parse(fs.Args()); err != nil {
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
