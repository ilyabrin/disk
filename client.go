package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const apiURL = "https://cloud-api.yandex.net/v1/disk/"

type HttpMethod string

const (
	GET    HttpMethod = http.MethodGet
	POST   HttpMethod = http.MethodPost
	PUT    HttpMethod = http.MethodPut
	PATCH  HttpMethod = http.MethodPatch
	DELETE HttpMethod = http.MethodDelete
)

// Client represents a Yandex.Disk API client.
//
// It encapsulates the necessary credentials and HTTP client configuration
// required to interact with the Yandex.Disk API endpoints.
type Client struct {
	// accessToken is the OAuth token used for authenticating with the Yandex.Disk API
	AccessToken string
	// httpClient is the underlying HTTP client used for making API requests
	HTTPClient *http.Client
}

// New creates a new Yandex.Disk API client.
//
// It initializes a client with the provided access token or retrieves it from
// the YANDEX_DISK_ACCESS_TOKEN environment variable if no token is provided.
//
// Parameters:
//   - token: Optional access token for Yandex.Disk API. If not provided,
//     the function will attempt to use the YANDEX_DISK_ACCESS_TOKEN environment variable.
//
// Returns:
//   - *Client: A configured client ready to interact with the Yandex.Disk API.
//   - error: An error if the access token is missing or empty, nil otherwise.
func New(token ...string) (*Client, error) {
	accessToken := ""
	if len(token) > 0 {
		accessToken = token[0]
	} else {
		accessToken = os.Getenv("YANDEX_DISK_ACCESS_TOKEN")
	}

	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	return &Client{
		AccessToken: accessToken,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// doRequest performs an HTTP request to the Yandex.Disk API and unmarshals the response into the specified type.
//
// Parameters:
//   - ctx: A context.Context for controlling the request lifecycle.
//   - method: The HTTP method to use for the request (GET, POST, etc.).
//   - resource: The API resource path to request.
//   - data: An io.Reader containing the request body, or nil for requests without a body.
//
// Returns:
//   - T: The unmarshaled response of the specified type.
//   - error: An error if the request could not be created, sent, or unmarshaled, or nil on success.
func doRequest[T any](ctx context.Context, c *Client, method HttpMethod, resource string, data io.Reader) (T, error) {
	var result T

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, string(method), apiURL+resource, data)
	if err != nil {
		return result, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil
}
