package disk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Context management and timeout handling implemented

const API_URL = "https://cloud-api.yandex.net/v1/disk/"

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	PATCH  HttpMethod = "PATCH"
	DELETE HttpMethod = "DELETE"
)

// ClientConfig holds configuration options for the Client
type ClientConfig struct {
	DefaultTimeout    time.Duration // Default timeout for requests
	MaxRetries        int           // Maximum number of retries (future use)
	EnableDebugLogging bool         // Enable debug logging (future use)
}

// DefaultClientConfig returns a ClientConfig with sensible defaults
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DefaultTimeout:    30 * time.Second,
		MaxRetries:        3,
		EnableDebugLogging: false,
	}
}

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	Logger      *log.Logger
	Config      *ClientConfig
}

// NewWithConfig creates a new Client with custom configuration
func NewWithConfig(config *ClientConfig, token ...string) (*Client, error) {
	if len(token) == 0 {
		envToken := os.Getenv("YANDEX_DISK_ACCESS_TOKEN")
		if envToken == "" {
			return nil, errors.New("provide yandex disk access token")
		}
		token = append(token, envToken)
	}

	if config == nil {
		config = DefaultClientConfig()
	}

	return &Client{
		AccessToken: token[0],
		HTTPClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
		Config: config,
	}, nil
}

// New(token ...string) fetch token from OS env var if has not direct defined
// Uses default configuration for backward compatibility
func New(token ...string) (*Client, error) {
	return NewWithConfig(nil, token...)
}

func (c *Client) doRequest(ctx context.Context, method HttpMethod, resource string, data io.Reader) (*http.Response, error) {
	// Ensure we have a proper context
	if ctx == nil {
		ctx = context.Background()
	}

	var resp *http.Response
	var err error
	var body io.Reader

	body = data

	// Use configurable timeout from client config if no deadline is set
	// This respects any existing context deadline while providing a fallback
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && c.Config != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Config.DefaultTimeout)
		defer cancel()
	} else if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// Fallback to HTTP client timeout if no config is available
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.HTTPClient.Timeout)
		defer cancel()
	}

	// Check if context is already cancelled before making the request
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	default:
		// Continue with request
	}

	if method == GET || method == DELETE {
		body = nil
	}

	req, err := http.NewRequestWithContext(ctx, string(method), API_URL+resource, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.AccessToken)

	if resp, err = c.HTTPClient.Do(req); err != nil {
		// Provide more context about the error
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request failed due to context: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, err
}

// WithTimeout creates a context with the specified timeout duration
func WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// WithDeadline creates a context with the specified deadline
func WithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), deadline)
}

// WithCancel creates a cancellable context
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// SetTimeout updates the default timeout for the client
func (c *Client) SetTimeout(timeout time.Duration) {
	if c.Config == nil {
		c.Config = DefaultClientConfig()
	}
	c.Config.DefaultTimeout = timeout
	c.HTTPClient.Timeout = timeout
}

// GetTimeout returns the current default timeout for the client
func (c *Client) GetTimeout() time.Duration {
	if c.Config != nil {
		return c.Config.DefaultTimeout
	}
	return c.HTTPClient.Timeout
}

// handleResponse provides centralized response handling with consistent error management
func (c *Client) handleResponse(resp *http.Response, expectedCodes []int) (*ErrorResponse, error) {
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200}
	}
	
	// Check if status code is expected
	for _, code := range expectedCodes {
		if resp.StatusCode == code {
			return nil, nil // Success
		}
	}
	
	// Handle error response
	var errorResponse ErrorResponse
	if resp.Body != nil {
		decoder := json.NewDecoder(resp.Body)
		if decodeErr := decoder.Decode(&errorResponse); decodeErr != nil {
			// If we can't decode the error response, create a generic one
			errorResponse = ErrorResponse{
				Error:       fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
				Description: fmt.Sprintf("Failed to decode error response: %v", decodeErr),
			}
		}
	} else {
		errorResponse = ErrorResponse{
			Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}
	
	return &errorResponse, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errorResponse.Error)
}

// safeDecodeJSON safely decodes JSON response with proper error handling for partial responses
func (c *Client) safeDecodeJSON(resp *http.Response, target interface{}) error {
	if resp.Body == nil {
		return fmt.Errorf("response body is nil")
	}
	
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(target); err != nil {
		// Check if this is a partial response or connection error
		if err.Error() == "EOF" {
			return fmt.Errorf("partial response received: connection may have been interrupted")
		}
		if err.Error() == "unexpected EOF" {
			return fmt.Errorf("incomplete response received: connection interrupted during transfer")
		}
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	return nil
}
