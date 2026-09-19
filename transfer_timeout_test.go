package disk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// slowBody streams a small body in chunks, taking `d` in total.
func slowBody(d time.Duration) http.HandlerFunc {
	const chunks = 5
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "50")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		for i := 0; i < chunks; i++ {
			_, _ = w.Write([]byte("0123456789"))
			if flusher != nil {
				flusher.Flush()
			}
			time.Sleep(d / chunks)
		}
	}
}

// newSlowClient wires a client whose DefaultTimeout is far shorter than the
// transfer takes, which is the shape of the real bug: the 30s default aborts
// any transfer slower than 30 seconds regardless of the caller's context.
func newSlowClient(t *testing.T, apiURL string) *Client {
	t.Helper()
	cfg := DefaultClientConfig()
	cfg.BaseURL = apiURL + "/"
	cfg.DefaultTimeout = 150 * time.Millisecond
	cfg.MaxRetries = 0
	client, err := NewWithConfig(cfg, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestDownloadOutlivesDefaultTimeout(t *testing.T) {
	files := httptest.NewServer(slowBody(750 * time.Millisecond))
	defer files.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Link{Href: files.URL, Method: http.MethodGet})
	}))
	defer api.Close()

	client := newSlowClient(t, api.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dst := filepath.Join(t.TempDir(), "out.bin")
	if err := client.DownloadFileToPath(ctx, "/big.bin", dst, &DownloadOptions{Overwrite: true}); err != nil {
		t.Fatalf("download aborted by the client timeout instead of honouring ctx: %v", err)
	}

	got, err := os.ReadFile(dst) // #nosec G304 -- path built by the test
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 50 {
		t.Errorf("expected the full 50-byte body, got %d bytes", len(got))
	}
}

func TestUploadOutlivesDefaultTimeout(t *testing.T) {
	var uploaded atomic.Int64

	upload := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Drain slowly so the request outlives DefaultTimeout.
		buf := make([]byte, 16)
		for {
			n, err := r.Body.Read(buf)
			uploaded.Add(int64(n))
			if err != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer upload.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "resources") && r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(ResourceUploadLink{Href: upload.URL, Method: http.MethodPut})
			return
		}
		_ = json.NewEncoder(w).Encode(Resource{Name: "big.bin", Path: "disk:/big.bin"})
	}))
	defer api.Close()

	client := newSlowClient(t, api.URL)

	src := filepath.Join(t.TempDir(), "big.bin")
	payload := strings.Repeat("x", 256)
	if err := os.WriteFile(src, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := client.UploadFileFromPath(ctx, src, "/big.bin", &UploadOptions{Overwrite: true}); err != nil {
		t.Fatalf("upload aborted by the client timeout instead of honouring ctx: %v", err)
	}
	if got := uploaded.Load(); got != int64(len(payload)) {
		t.Errorf("expected the server to receive %d bytes, got %d", len(payload), got)
	}
}

// A caller's own deadline must still win over the transfer fallback.
func TestTransferHonoursCallerDeadline(t *testing.T) {
	files := httptest.NewServer(slowBody(2 * time.Second))
	defer files.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Link{Href: files.URL, Method: http.MethodGet})
	}))
	defer api.Close()

	client := newSlowClient(t, api.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	dst := filepath.Join(t.TempDir(), "out.bin")
	err := client.DownloadFileToPath(ctx, "/big.bin", dst, &DownloadOptions{Overwrite: true})
	if err == nil {
		t.Fatal("expected the caller's short deadline to abort the download")
	}
}

func TestTransferClientHasNoTimeout(t *testing.T) {
	client, err := New("token")
	if err != nil {
		t.Fatal(err)
	}
	if client.HTTPClient.Timeout == 0 {
		t.Fatal("precondition: the default client is expected to carry a timeout")
	}
	if got := client.transferClient().Timeout; got != 0 {
		t.Errorf("transfer client must not carry a wall-clock timeout, got %v", got)
	}
	if client.transferClient().Transport != client.HTTPClient.Transport {
		t.Error("transfer client should reuse the shared transport")
	}
}

func TestTransferContext(t *testing.T) {
	// No caller deadline: the fallback applies.
	ctx, cancel := transferContext(context.Background())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected a fallback deadline")
	}
	if remaining := time.Until(deadline); remaining < 29*time.Minute {
		t.Errorf("fallback deadline too short: %v", remaining)
	}

	// Caller deadline: left untouched.
	own, cancelOwn := context.WithTimeout(context.Background(), time.Minute)
	defer cancelOwn()
	got, cancelGot := transferContext(own)
	defer cancelGot()
	if got != own {
		t.Error("expected the caller's context to be returned unchanged")
	}
}
