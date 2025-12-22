package webauthn

import "flag"

type Options struct {
	Port string
	Host string

	AllowedOrigins []string

	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Port, "p", "8080", "Server port (default: 8080)")
	fs.StringVar(&opts.Port, "port", "8080", "Server port (default: 8080)")
	fs.StringVar(&opts.Host, "H", "0.0.0.0", "Server host (default: '0.0.0.0')")
	fs.StringVar(&opts.Host, "host", "0.0.0.0", "Server host (default: '0.0.0.0')")
	fs.Func("c", "CORS Allowed Origins (comma separated)", func(s string) error {
		if s == "" {
			return nil
		}
		opts.AllowedOrigins = append(opts.AllowedOrigins, s)
		return nil
	})
	fs.StringVar(&opts.RPDisplayName, "n", "GoChat", "Relying Party Display Name (default: 'GoChat')")
	fs.StringVar(&opts.RPID, "i", "localhost", "Relying Party ID (default: 'localhost')")
	fs.Func("o", "Relying Party Origins (comma separated)", func(s string) error {
		if s == "" {
			return nil
		}
		opts.RPOrigins = append(opts.RPOrigins, s)
		return nil
	})

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return &opts, nil
}
