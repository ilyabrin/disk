package disk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
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
	DefaultTimeout time.Duration // Default timeout for requests
	// MaxRetries bounds how many extra attempts a retryable request gets.
	// Only requests without a body (GET/DELETE/PUT with query parameters) are
	// retried, and only on connection errors, 429 and 5xx responses.
	MaxRetries         int
	RetryBackoff       time.Duration // Base delay between retries; doubles per attempt
	EnableDebugLogging bool          // Enable debug logging
	Logger             *LoggerConfig // Logger configuration
	// BaseURL overrides the Yandex Disk API endpoint. Useful for tests and
	// proxies; defaults to API_URL. Must end with a slash.
	BaseURL string
}

// DefaultClientConfig returns a ClientConfig with sensible defaults
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DefaultTimeout:     30 * time.Second,
		MaxRetries:         3,
		RetryBackoff:       200 * time.Millisecond,
		EnableDebugLogging: false,
		Logger:             DefaultLoggerConfig(),
		BaseURL:            API_URL,
	}
}

// baseURL returns the API endpoint this client talks to.
func (c *Client) baseURL() string {
	if c.Config != nil && c.Config.BaseURL != "" {
		return c.Config.BaseURL
	}
	return API_URL
}

type Client struct {
	AccessToken string
	HTTPClient  *http.Client
	Logger      *DiskLogger
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

	// Validate and sanitize token
	sanitizedToken := strings.TrimSpace(token[0])
	if sanitizedToken == "" {
		return nil, errors.New("access token cannot be empty")
	}

	// Initialize logger
	logger := NewLogger(config.Logger)
	if config.EnableDebugLogging {
		logger.SetLevel(DEBUG)
		logger.SetVerbose(true)
	}

	// Create HTTP client with secure TLS configuration
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			},
		},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		AccessToken: sanitizedToken,
		HTTPClient: &http.Client{
			Timeout:   config.DefaultTimeout,
			Transport: transport,
		},
		Config: config,
		Logger: logger,
	}, nil
}

// New(token ...string) fetch token from OS env var if has not direct defined
// Uses default configuration for backward compatibility
func New(token ...string) (*Client, error) {
	return NewWithConfig(nil, token...)
}

func (c *Client) doRequest(ctx context.Context, method HttpMethod, resource string, data io.Reader) (*http.Response, error) {
	startTime := time.Now()

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
		c.Logger.LogError("doRequest", ctx.Err())
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	default:
		// Continue with request
	}

	if method == GET || method == DELETE {
		body = nil
	} else if data != nil {
		// Limit request body size to prevent memory exhaustion
		body = io.LimitReader(data, 100*1024*1024) // 100MB limit
	}

	requestURL := c.baseURL() + resource
	req, err := http.NewRequestWithContext(ctx, string(method), requestURL, body)
	if err != nil {
		c.Logger.LogError("create request", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+c.AccessToken)

	// Log request details
	if c.Logger != nil {
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		c.Logger.LogRequest(string(method), requestURL, headers)
	}

	resp, err = c.doWithRetries(ctx, req, body == nil)
	if err != nil {
		c.Logger.LogError("execute request", err)

		// Provide more context about the error
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request failed due to context: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Log response details
	if c.Logger != nil {
		duration := time.Since(startTime)
		contentLength := resp.ContentLength
		if contentLength == -1 {
			contentLength = 0
		}
		c.Logger.LogResponse(resp.StatusCode, contentLength, duration)
	}

	return resp, err
}

// doWithRetries executes req, retrying transient failures when the request can
// safely be replayed. Only bodyless requests are retryable: an io.Reader body
// cannot be rewound for a second attempt.
func (c *Client) doWithRetries(ctx context.Context, req *http.Request, retryable bool) (*http.Response, error) {
	maxRetries := 0
	backoff := 200 * time.Millisecond
	if retryable && c.Config != nil && c.Config.MaxRetries > 0 {
		maxRetries = c.Config.MaxRetries
		if c.Config.RetryBackoff > 0 {
			backoff = c.Config.RetryBackoff
		}
	}

	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		resp, err = c.HTTPClient.Do(req)

		if attempt >= maxRetries || !shouldRetry(resp, err) {
			return resp, err
		}

		if resp != nil {
			// The body has to be drained and closed before the connection can
			// be reused for the next attempt.
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
		}

		c.Logger.Debug("Retrying request %s %s (attempt %d/%d)", req.Method, req.URL.Path, attempt+1, maxRetries)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff << attempt):
		}
	}
}

// shouldRetry reports whether a failed attempt is worth repeating.
func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
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

// SetLogLevel sets the minimum log level for the client
func (c *Client) SetLogLevel(level LogLevel) {
	if c.Logger != nil {
		c.Logger.SetLevel(level)
	}
}

// SetVerbose enables or disables verbose logging
func (c *Client) SetVerbose(verbose bool) {
	if c.Logger != nil {
		c.Logger.SetVerbose(verbose)
	}
	if c.Config != nil {
		c.Config.EnableDebugLogging = verbose
	}
}

// SetLogOutput changes the log output destination
func (c *Client) SetLogOutput(output io.Writer) {
	if c.Logger != nil {
		c.Logger.SetOutput(output)
	}
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

// requestJSON performs an API request and decodes a successful response into T.
//
// okCodes lists the HTTP status codes that count as success (defaults to 200).
// Any other status is decoded into an *ErrorResponse, which is what every
// resource/public endpoint in this package returns to the caller.
func requestJSON[T any](ctx context.Context, c *Client, method HttpMethod, endpoint string, body io.Reader, okCodes ...int) (*T, *ErrorResponse) {
	if len(okCodes) == 0 {
		okCodes = []int{http.StatusOK}
	}

	resp, err := c.doRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if !slices.Contains(okCodes, resp.StatusCode) {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, &errorResponse
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode response: %v", err)}
	}
	return &result, nil
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
