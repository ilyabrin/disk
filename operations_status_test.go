package disk

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestOperationIDFromHref(t *testing.T) {
	cases := map[string]string{
		"https://cloud-api.yandex.net/v1/disk/operations/123abc":  "123abc",
		"https://cloud-api.yandex.net/v1/disk/operations/123abc/": "123abc",
		"123abc": "123abc",
		"":       "",
	}

	for href, want := range cases {
		if got := OperationIDFromHref(href); got != want {
			t.Errorf("OperationIDFromHref(%q) = %q, want %q", href, got, want)
		}
	}
}

func TestGetOperationStatus(t *testing.T) {
	t.Run("returns the operation", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("operation_id"); got != "123abc" {
				t.Errorf("operation_id = %q, want %q", got, "123abc")
			}
			_, _ = w.Write([]byte(`{"status":"success"}`))
		})

		operation, err := client.GetOperationStatus(context.Background(),
			"https://cloud-api.yandex.net/v1/disk/operations/123abc")
		if err != nil {
			t.Fatal(err)
		}
		if operation.Status != "success" {
			t.Errorf("status = %q, want %q", operation.Status, "success")
		}
	})

	t.Run("rejects an empty id", func(t *testing.T) {
		client, _ := New("token")
		if _, err := client.GetOperationStatus(context.Background(), ""); err == nil {
			t.Error("expected an error for an empty operation id")
		}
	})

	t.Run("surfaces API errors", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"NotFound","description":"no such operation"}`))
		})
		client.Config.MaxRetries = 0

		if _, err := client.GetOperationStatus(context.Background(), "missing"); err == nil {
			t.Error("expected an error for a missing operation")
		}
	})
}

func TestWaitForBatchOperationPolls(t *testing.T) {
	var calls int32
	client := mockedHttpClient(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) < 2 {
			_, _ = w.Write([]byte(`{"status":"in-progress"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"success"}`))
	})

	status := &BatchOperationStatus{
		Results: []*BatchOperationResult{
			{Path: "/a", Operation: "copy", Link: &Link{Href: "https://cloud-api.yandex.net/v1/disk/operations/op1"}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.WaitForBatchOperation(ctx, status, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&calls); got < 2 {
		t.Errorf("expected the operation to be polled more than once, got %d calls", got)
	}
}
