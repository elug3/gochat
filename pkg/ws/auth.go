package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpAuthenticator struct {
	AuthUrl string
}

func NewHttpAuthenticator(authUrl string) (*HttpAuthenticator, error) {
	return &HttpAuthenticator{
		AuthUrl: authUrl,
	}, nil
}

func (auth *HttpAuthenticator) ValidateWsToken(ctx context.Context, wsToken string) (int32, error) {
	u := "http://" + auth.AuthUrl + "/ws?token=" + wsToken
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, err
	}

	var result struct {
		UserId int32 `json:"user_id"`
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("authentication failed with status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return result.UserId, nil
}
