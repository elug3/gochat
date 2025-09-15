package httpclient

import (
	"flag"
	"net/http"
	"os"
)

type RequestOption func(*RequestConfig)

func WithBaseUrl(baseUrl string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.BaseUrl = baseUrl
	}
}

func WithHttpClient(client *http.Client) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.HttpClient = client
	}
}

func WithApiToken(token string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Header.Add("Authorization", "Bearer "+token)
	}
}

func WithHeader(key, value string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Header.Add(key, value)
	}
}

func DefaultRequestOptions() []RequestOption {
	var opts []RequestOption

	if baseUrl, ok := os.LookupEnv("GOCHAT_BASE_URL"); ok {
		opts = append(opts, WithBaseUrl(baseUrl))
	}
	if token, ok := os.LookupEnv("GOCHAT_API_TOKEN"); ok {
		opts = append(opts, WithApiToken(token))
	}

	return opts
}

func ConfigureOptions(fs *flag.FlagSet, args []string) ([]RequestOption, error) {
	var (
		baseUrl  string
		apiToken string
	)

	fs.StringVar(&baseUrl, "base-url", "", "Base URL for the API")
	fs.StringVar(&apiToken, "api-token", "", "API token for authentication")

	fs.Parse(args)

	if baseUrl == "" {
		baseUrl = os.Getenv("GOCHAT_BASE_URL")
	}
	if apiToken == "" {
		apiToken = os.Getenv("GOCHAT_API_TOKEN")
	}

	var opts []RequestOption
	if baseUrl != "" {
		opts = append(opts, WithBaseUrl(baseUrl))
	}
	if apiToken != "" {
		opts = append(opts, WithApiToken(apiToken))
	}

	return opts, nil
}
