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

type HTTPHeaders map[string]string
type QueryParams map[string]string

type Metadata map[string]map[string]string

type Client struct {
	accessToken string
	HTTPClient  *http.Client
	logger      *log.Logger

	apiURL string
	reqURL string // for easy testing
}

func New(token string) *Client {
	if len(token) == 0 {
		return nil
	}

	return &Client{
		accessToken: token,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		logger:      &log.Logger{},
		apiURL:      API_URL,
		reqURL:      "",
	}
}

func (c *Client) doRequest(ctx context.Context, method string, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	var resp *http.Response
	var err error
	var data io.Reader

	// TODO
	// ctx, cancel := context.WithCancel(ctx)

	data = body

	if method == "GET" || method == "DELETE" {
		data = nil
	}

	req, err := http.NewRequestWithContext(ctx, method, resource, data)
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

	c.reqURL = req.URL.String()

	if resp, err = c.HTTPClient.Do(req); err != nil {
		// c.logger.Fatal("error response", err)
		return nil, err
	}

	return resp, err
}

func (c *Client) ReqURL() string {
	return c.reqURL
}

func (c *Client) ApiURL() string {
	return c.apiURL
}

func (c *Client) get(ctx context.Context, resource string, params *QueryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, resource, nil, nil, params)
}

func (c *Client) post(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, resource, body, headers, params)
}

func (c *Client) patch(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPatch, resource, body, headers, params)
}

func (c *Client) put(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, resource, body, headers, params)
}

func (c *Client) delete(ctx context.Context, resource string, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, resource, nil, headers, params)
}
