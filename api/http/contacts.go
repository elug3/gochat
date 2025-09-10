package httpclient

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/go-querystring/query"
)

type ContactsClient struct {
	BaseUrl string
}

func NewContactsClient(baseUrl string) *ContactsClient {
	return &ContactsClient{BaseUrl: baseUrl}
}

type AccessRequestParams struct {
	UserId   int    `url:"user_id"`
	ChatId   int    `url:"chat_id"`
	TargetId int    `url:"target_id"`
	Action   string `url:"action"`
}

func (c *ContactsClient) Can(params AccessRequestParams) (bool, string, error) {
	reqUrl := "http://" + c.BaseUrl + "/can"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return false, "", err
	}

	v, err := query.Values(params)
	if err != nil {
		return false, "", err
	}
	req.URL.RawQuery = v.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var res struct {
		Can    bool   `json:"can"`
		Action string `json:"action"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, "", err
	}
	if res.Error != "" {
		err = errors.New(res.Error)
	}

	return res.Can, res.Action, err

}
