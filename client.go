package disk

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

// todo: add context cancellation

const API_URL = "https://cloud-api.yandex.net/v1/disk/"

type HTTPHeaders map[string]string
type QueryParams map[string]string

type Metadata map[string]map[string]string

type service struct{ client *Client }
type Client struct {
	accessToken string
	HTTPClient  *http.Client
	logger      *log.Logger

	apiURL string
	reqURL string // for easy testing

	common service

	Disk      *DiskService
	Trash     *TrashService
	Public    *PublicService
	Resources *ResourceService
	Operation *OperationService
}

func New(token string) *Client {
	if len(token) == 0 {
		return nil
	}

	c := &Client{
		accessToken: token,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		logger:      &log.Logger{},
		apiURL:      API_URL,
		reqURL:      "",
	}
	c.common.client = c

	c.Disk = (*DiskService)(&c.common)
	c.Trash = (*TrashService)(&c.common)
	c.Public = (*PublicService)(&c.common)
	c.Resources = (*ResourceService)(&c.common)
	c.Operation = (*OperationService)(&c.common)

	return c
}

func (c *Client) do(ctx context.Context, method string, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	var resp *http.Response
	var err error
	var data io.Reader

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
	return c.do(ctx, http.MethodGet, resource, nil, nil, params)
}

func (c *Client) post(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, resource, body, headers, params)
}

func (c *Client) patch(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.do(ctx, http.MethodPatch, resource, body, headers, params)
}

func (c *Client) put(ctx context.Context, resource string, body io.Reader, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, resource, body, headers, params)
}

func (c *Client) delete(ctx context.Context, resource string, headers *HTTPHeaders, params *QueryParams) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, resource, nil, headers, params)
}
