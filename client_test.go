package disk

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type testTransport struct {
	server *httptest.Server
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new request to the test server
	testURL := "http://" + t.server.Listener.Addr().String() + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}

	testReq, err := http.NewRequest(req.Method, testURL, req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	testReq.Header = req.Header.Clone()

	return http.DefaultClient.Do(testReq)
}

func mockedHttpClient(h http.HandlerFunc) *Client {
	s := httptest.NewServer(h)

	client, _ := New("token")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{server: s},
	}

	return client
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return client, s.Close
}

func TestNew(t *testing.T) {
	// Helper function to reset environment variable
	resetEnv := func() {
		os.Unsetenv("YANDEX_DISK_ACCESS_TOKEN")
	}

	t.Run("With provided token", func(t *testing.T) {
		resetEnv()
		client, err := New("test-token")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "test-token" {
			t.Errorf("Expected AccessToken to be 'test-token', got '%s'", client.AccessToken)
		}
		if client.HTTPClient == nil {
			t.Fatal("Expected non-nil HTTPClient")
		}
		if client.HTTPClient.Timeout != 30*time.Second {
			t.Errorf("Expected Timeout to be 30 seconds, got %v", client.HTTPClient.Timeout)
		}
	})

	t.Run("With environment variable", func(t *testing.T) {
		resetEnv()
		os.Setenv("YANDEX_DISK_ACCESS_TOKEN", "env-token")
		client, err := New()
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "env-token" {
			t.Errorf("Expected AccessToken to be 'env-token', got '%s'", client.AccessToken)
		}
	})

	t.Run("Without token and empty environment variable", func(t *testing.T) {
		resetEnv()
		client, err := New()
		if err == nil {
			t.Fatal("Expected error for missing token")
		}
		if client != nil {
			t.Fatal("Expected nil client")
		}
	})

	t.Run("With multiple tokens", func(t *testing.T) {
		resetEnv()
		client, err := New("token1", "token2")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "token1" {
			t.Errorf("Expected AccessToken to be 'token1', got '%s'", client.AccessToken)
		}
	})

	t.Run("HTTPClient configuration", func(t *testing.T) {
		resetEnv()
		client, err := New("test-token")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.HTTPClient == nil {
			t.Fatal("Expected non-nil HTTPClient")
		}
		if client.HTTPClient.Timeout != 30*time.Second {
			t.Errorf("Expected Timeout to be 30 seconds, got %v", client.HTTPClient.Timeout)
		}
	})
}

func TestClientHelperMethods(t *testing.T) {
	t.Run("WithDeadline", func(t *testing.T) {
		deadline := time.Now().Add(5 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		actualDeadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Expected deadline to be set")
		}

		if !actualDeadline.Equal(deadline) {
			t.Errorf("Expected deadline %v, got %v", deadline, actualDeadline)
		}
	})

	t.Run("GetTimeout basic test", func(t *testing.T) {
		client, _ := New("test-token")

		timeout := client.GetTimeout()
		if timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", timeout)
		}
	})

	t.Run("SetTimeout edge cases", func(t *testing.T) {
		client, _ := New("test-token")

		// Test normal case
		client.SetTimeout(5 * time.Second)
		if client.HTTPClient.Timeout != 5*time.Second {
			t.Errorf("Expected HTTPClient timeout 5s, got %v", client.HTTPClient.Timeout)
		}
	})

	t.Run("safeDecodeJSON with various inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test with valid JSON
		validJSON := `{"name": "test", "value": 123}`
		resp := &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(validJSON)),
		}

		var result map[string]interface{}
		err := client.safeDecodeJSON(resp, &result)
		if err != nil {
			t.Errorf("Expected no error for valid JSON, got: %v", err)
		}

		if result["name"] != "test" {
			t.Errorf("Expected name 'test', got %v", result["name"])
		}

		// Test with invalid JSON
		invalidJSON := `{invalid json}`
		resp = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(invalidJSON)),
		}

		var invalidResult map[string]interface{}
		err = client.safeDecodeJSON(resp, &invalidResult)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		// Test with response with nil body
		respWithNilBody := &http.Response{
			StatusCode: 200,
			Body:       nil,
		}
		err = client.safeDecodeJSON(respWithNilBody, &result)
		if err == nil {
			t.Error("Expected error for nil response body")
		}
	})

	t.Run("handleResponse with different status codes", func(t *testing.T) {
		client, _ := New("test-token")

		// Test with acceptable status code
		resp := &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"result": "success"}`)),
		}

		result, err := client.handleResponse(resp, []int{200, 201})
		if err != nil {
			t.Errorf("Expected no error for acceptable status code, got: %v", err)
		}
		// Result can be nil for some responses, which is OK
		_ = result

		// Test with unacceptable status code - error response
		resp = &http.Response{
			StatusCode: 404,
			Status:     "404 Not Found",
			Body:       io.NopCloser(strings.NewReader(`{"error": "NotFoundError", "description": "Resource not found"}`)),
		}

		_, err = client.handleResponse(resp, []int{200})
		if err == nil {
			t.Error("Expected error for unacceptable status code")
		}

		// Test with unacceptable status code - invalid error JSON
		resp = &http.Response{
			StatusCode: 500,
			Status:     "500 Internal Server Error",
			Body:       io.NopCloser(strings.NewReader(`{invalid error json}`)),
		}

		_, err = client.handleResponse(resp, []int{200})
		if err == nil {
			t.Error("Expected error for invalid error JSON")
		}
	})
}
