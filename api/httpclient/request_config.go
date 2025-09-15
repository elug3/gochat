package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestConfig struct {
	Method     string
	HttpClient *http.Client
	BaseUrl    string
	// Full URL including baseUrl + path + query parameters
	Endpoint string
	Header   http.Header
	Body     io.Reader
}

func NewRequestConfig(method, path string, body io.Reader, opts ...RequestOption) (*RequestConfig, error) {
	var cfg RequestConfig
	cfg.Header = make(http.Header)

	for _, opt := range opts {
		opt(&cfg)
	}

	cfg.Method = method
	cfg.Body = body
	if cfg.HttpClient == nil {
		cfg.HttpClient = http.DefaultClient
	}
	if cfg.BaseUrl == "" {
		cfg.BaseUrl = "localhost:8080"
	}
	cfg.Endpoint = "http://" + cfg.BaseUrl + path

	return &cfg, nil
}

func (cfg *RequestConfig) Do(ctx context.Context, dst interface{}) error {
	req, err := http.NewRequestWithContext(ctx, cfg.Method, cfg.Endpoint, cfg.Body)
	if err != nil {
		return err
	}
	req.Header = cfg.Header

	resp, err := cfg.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 400 {
		switch resp.StatusCode {
		case 401:
			return fmt.Errorf("unauthorized (401): please check your API token")
		case 403:
			return fmt.Errorf("forbidden (403): you don't have permission to access this resource")
		case 404:
			return fmt.Errorf("page not found (404): the requested resource does not exist at %s", cfg.Endpoint)
		case 500:
			return fmt.Errorf("internal server error (500): something went wrong on the server side")
		default:
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	if err = json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return err
	}

	return nil
}
