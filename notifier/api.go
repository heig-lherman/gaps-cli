package notifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	BaseUrl string
	ApiKey  string

	httpClient *http.Client
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient(apiUrl string, apiKey string) *Client {
	return &Client{
		BaseUrl: apiUrl,
		ApiKey:  apiKey,

		httpClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ApiKey))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}

		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	if v == nil {
		return nil
	}

	if err = json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}

	return nil
}
