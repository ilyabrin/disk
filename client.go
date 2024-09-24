package disk

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// todo: add context cancellation

const API_URL = "https://cloud-api.yandex.net/v1/disk/"

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	PATCH  HttpMethod = "PATCH"
	DELETE HttpMethod = "DELETE"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	Logger      *log.Logger
}

// New(token ...string) fetch token from OS env var if has not direct defined
func New(token ...string) *Client {
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
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method HttpMethod, resource string, data io.Reader) (*http.Response, error) {

	var resp *http.Response
	var err error
	var body io.Reader

	body = data

	// todo: make time parameterized, not const
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
	defer cancel()

	if method == GET || method == DELETE {
		body = nil
	}

	req, err := http.NewRequestWithContext(ctx, string(method), API_URL+resource, body)
	if err != nil {
		c.Logger.Fatal("error request", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.AccessToken)

	if resp, err = c.HTTPClient.Do(req); err != nil {
		c.Logger.Fatal("error response", err)
		return nil, err
	}

	return resp, err
}
