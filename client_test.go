package disk

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func mockedHttpClient(h http.HandlerFunc) *Client {
	httpClient, _ := testingHTTPClient(h)

	client := *New("token")
	client.HTTPClient = httpClient

	return &client
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
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
		client := New("test-token")
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "test-token" {
			t.Errorf("Expected AccessToken to be 'test-token', got '%s'", client.AccessToken)
		}
		if client.HTTPClient == nil {
			t.Fatal("Expected non-nil HTTPClient")
		}
		if client.HTTPClient.Timeout != 10*time.Second {
			t.Errorf("Expected Timeout to be 10 seconds, got %v", client.HTTPClient.Timeout)
		}
	})

	t.Run("With environment variable", func(t *testing.T) {
		resetEnv()
		os.Setenv("YANDEX_DISK_ACCESS_TOKEN", "env-token")
		client := New()
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "env-token" {
			t.Errorf("Expected AccessToken to be 'env-token', got '%s'", client.AccessToken)
		}
	})

	t.Run("Without token and empty environment variable", func(t *testing.T) {
		resetEnv()
		client := New()
		if client != nil {
			t.Fatal("Expected nil client")
		}
	})

	t.Run("With multiple tokens", func(t *testing.T) {
		resetEnv()
		client := New("token1", "token2")
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.AccessToken != "token1" {
			t.Errorf("Expected AccessToken to be 'token1', got '%s'", client.AccessToken)
		}
	})

	t.Run("HTTPClient configuration", func(t *testing.T) {
		resetEnv()
		client := New("test-token")
		if client == nil {
			t.Fatal("Expected non-nil client")
		}
		if client.HTTPClient == nil {
			t.Fatal("Expected non-nil HTTPClient")
		}
		if client.HTTPClient.Timeout != 10*time.Second {
			t.Errorf("Expected Timeout to be 10 seconds, got %v", client.HTTPClient.Timeout)
		}
	})
}
