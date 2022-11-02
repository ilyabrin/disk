package disk

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

// todo: add context cancellation

// TODO: remove
const API_URL = "https://cloud-api.yandex.net/v1/disk/"

type httpHeaders map[string]string
type queryParams map[string]string

type metadata map[string]map[string]string

type Client struct {
	accessToken string
	httpClient  *http.Client
	logger      *log.Logger

	api_url string
	req_url string // for easy testing
}

func New(token string) *Client {

	if len(token) == 0 {
		return nil
	}

	return &Client{
		accessToken: token,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		logger:      &log.Logger{},
		api_url:     API_URL,
		req_url:     "",
	}
}

func (c *Client) doRequest(ctx context.Context, method string, resource string, body io.Reader, headers *httpHeaders, params *queryParams) (*http.Response, error) {

	var resp *http.Response
	var err error
	var data io.Reader

	// TODO
	// ctx, cancel := context.WithCancel(ctx)

	data = body

	if method == "GET" || method == "DELETE" {
		data = nil
	}

	req, err := http.NewRequestWithContext(ctx, string(method), resource, data)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.accessToken)

	if headers != nil {
		for key, value := range *headers {
			req.Header.Add(key, value)
		}
	}

	if params != nil {
		q := req.URL.Query()
		for key, value := range *params {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	c.req_url = req.URL.String()

	if resp, err = c.httpClient.Do(req); err != nil {
		// c.logger.Fatal("error response", err)
		return nil, err
	}

	return resp, err
}

func (c *Client) get(ctx context.Context, resource string, headers *httpHeaders, params *queryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, resource, nil, headers, params)
}

func (c *Client) post(ctx context.Context, resource string, body io.Reader, headers *httpHeaders, params *queryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, resource, nil, headers, params)
}

func (c *Client) patch(ctx context.Context, resource string, body io.Reader, headers *httpHeaders, params *queryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPatch, resource, nil, headers, params)
}

func (c *Client) put(ctx context.Context, resource string, body io.Reader, headers *httpHeaders, params *queryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, resource, nil, headers, params)
}

func (c *Client) delete(ctx context.Context, resource string, headers *httpHeaders, params *queryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, resource, nil, headers, params)
}
