package disk

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryOnServerError(t *testing.T) {
	t.Run("retries 5xx until success", func(t *testing.T) {
		var calls int32
		client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
			if atomic.AddInt32(&calls, 1) < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"file.txt"}`))
		})
		client.Config.RetryBackoff = time.Millisecond

		resource, errResp := client.GetMetadata(context.Background(), "/file.txt")
		if errResp != nil {
			t.Fatalf("expected success after retries, got %v", errResp)
		}
		if resource.Name != "file.txt" {
			t.Errorf("expected file.txt, got %q", resource.Name)
		}
		if got := atomic.LoadInt32(&calls); got != 3 {
			t.Errorf("expected 3 attempts, got %d", got)
		}
	})

	t.Run("gives up after MaxRetries", func(t *testing.T) {
		var calls int32
		client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"InternalError"}`))
		})
		client.Config.MaxRetries = 2
		client.Config.RetryBackoff = time.Millisecond

		if _, errResp := client.GetMetadata(context.Background(), "/file.txt"); errResp == nil {
			t.Fatal("expected an error response")
		}
		if got := atomic.LoadInt32(&calls); got != 3 {
			t.Errorf("expected 1 initial attempt + 2 retries, got %d", got)
		}
	})

	t.Run("does not retry 4xx", func(t *testing.T) {
		var calls int32
		client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"NotFound"}`))
		})
		client.Config.RetryBackoff = time.Millisecond

		if _, errResp := client.GetMetadata(context.Background(), "/missing"); errResp == nil {
			t.Fatal("expected an error response")
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Errorf("expected exactly 1 attempt, got %d", got)
		}
	})

	t.Run("does not retry requests with a body", func(t *testing.T) {
		var calls int32
		client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"BadGateway"}`))
		})
		client.Config.RetryBackoff = time.Millisecond

		props := map[string]map[string]string{"custom_properties": {"k": "v"}}
		if _, errResp := client.UpdateMetadata(context.Background(), "/file.txt", props); errResp == nil {
			t.Fatal("expected an error response")
		}
		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Errorf("expected exactly 1 attempt for a request with a body, got %d", got)
		}
	})
}

func TestBaseURLOverride(t *testing.T) {
	client, err := NewWithConfig(&ClientConfig{
		DefaultTimeout: 5 * time.Second,
		Logger:         DefaultLoggerConfig(),
		BaseURL:        "https://example.invalid/v1/disk/",
	}, "token")
	if err != nil {
		t.Fatal(err)
	}

	if got := client.baseURL(); got != "https://example.invalid/v1/disk/" {
		t.Errorf("expected the configured base URL, got %q", got)
	}

	plain, _ := New("token")
	if got := plain.baseURL(); got != API_URL {
		t.Errorf("expected the default base URL, got %q", got)
	}
}
