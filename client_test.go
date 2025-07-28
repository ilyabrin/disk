package disk

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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
