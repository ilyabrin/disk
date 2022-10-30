package disk

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// todo: add context cancellation

// TODO: remove
const API_URL = "https://cloud-api.yandex.net/v1/disk/"

type Method string
type optional_params map[string]string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETE"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	Logger      *log.Logger

	api_url string
}

// New(token ...string) fetch token from OS env var if has not direct defined
func New(token ...string) *Client {
	// # TODO: remove env token declaration
	if len(token) == 0 {
		envToken := os.Getenv("YANDEX_DISK_ACCESS_TOKEN")
		if envToken == "" {
			return nil
		}
		token = append(token, envToken)
	}

	return &Client{
		AccessToken: token[0],
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		api_url: API_URL,
	}
}

func (c *Client) doRequest(ctx context.Context, method Method, resource string, body io.Reader, params *optional_params) (*http.Response, error) {

	var resp *http.Response
	var err error
	var data io.Reader

	// ctx, cancel := context.WithCancel(ctx)

	data = body

	if method == GET || method == DELETE {
		data = nil
	}

	req, err := http.NewRequestWithContext(ctx, string(method), resource, data)
	if err != nil {
		panic(err)
	}
	// req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.AccessToken)

	if params != nil {
		q := req.URL.Query()
		for key, value := range *params {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// TODO: remove
	fmt.Println("URL Query:", req.URL.String())

	if resp, err = c.HTTPClient.Do(req); err != nil {
		c.Logger.Fatal("error response", err)
		return nil, err
	}

	return resp, err
}
