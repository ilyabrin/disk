package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// todo: add context cancellation

const API_URL = "https://cloud-api.yandex.net/v1/"

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	Logger      *log.Logger
}

func New(token ...string) *Client {
	if len(token) < 1 {
		token = append(token, os.Getenv("YANDEX_DISK_ACCESS_TOKEN"))
	}

	if len(token) <= 1 {
		return nil
	}
	return &Client{
		AccessToken: token[0],
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method string, resource string, body io.Reader) (*http.Response, error) {

	var resp *http.Response
	var err error
	var data io.Reader

	// ctx, cancel := context.WithCancel(ctx)
	if method == GET || method == DELETE {
		data = nil
	}
	data = body

	req, err := http.NewRequestWithContext(ctx, method, API_URL+resource, data)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.AccessToken)

	if resp, err = c.HTTPClient.Do(req); err != nil {
		c.Logger.Fatal("error response", err)
		return nil, err
	}

	return resp, err
}
