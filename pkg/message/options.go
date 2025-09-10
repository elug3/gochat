package message

import "flag"

type Options struct {
	Port               string
	Host               string
	SaveDir            string
	NoSave             bool
	NatsUrl            string
	ContactsServiceUrl string
}

func ConfigureOptions(fs *flag.FlagSet, args []string) (*Options, error) {
	var opts Options

	fs.StringVar(&opts.Port, "port", "8080", "Port to run the server on")
	fs.StringVar(&opts.Host, "host", "localhost", "Host to run the server on")
	fs.StringVar(&opts.SaveDir, "save-dir", "./", "Directory to save data")
	fs.BoolVar(&opts.NoSave, "no-save", false, "Run in no-save mode (data will not be persisted)")
	fs.StringVar(&opts.NatsUrl, "nats-url", "nats://localhost:4222", "URL for the NATS server")
	fs.StringVar(&opts.ContactsServiceUrl, "contacts-url", "localhost:8002", "URL for the contacts service")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &opts, nil
}
