package disk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// dirServer records every path PUT to /resources and can pretend that some of
// them already exist.
type dirServer struct {
	mu       sync.Mutex
	created  []string
	existing map[string]bool
}

func (d *dirServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")

		d.mu.Lock()
		exists := d.existing[path]
		if !exists {
			d.created = append(d.created, path)
		}
		d.mu.Unlock()

		if exists {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Message:     "Resource already exists",
				Error:       "DiskPathPointsToExistentDirectoryError",
				Description: "Specified path already exists",
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Link{Href: "https://example.invalid" + path, Method: http.MethodGet})
	}
}

func newDirClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	cfg := DefaultClientConfig()
	cfg.BaseURL = srv.URL + "/"
	cfg.MaxRetries = 0
	client, err := NewWithConfig(cfg, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestCreateDirAllCreatesEverySegment(t *testing.T) {
	d := &dirServer{existing: map[string]bool{}}
	srv := httptest.NewServer(d.handler())
	defer srv.Close()

	client := newDirClient(t, srv)

	if errResp := client.CreateDirAll(context.Background(), "/newDir/subDir/anotherDir"); errResp != nil {
		t.Fatalf("unexpected error: %v", errResp.Error)
	}

	want := []string{"/newDir", "/newDir/subDir", "/newDir/subDir/anotherDir"}
	if len(d.created) != len(want) {
		t.Fatalf("expected %d requests, got %v", len(want), d.created)
	}
	for i, path := range want {
		if d.created[i] != path {
			t.Errorf("request %d: expected %q, got %q", i, path, d.created[i])
		}
	}
}

func TestCreateDirAllSkipsExistingParents(t *testing.T) {
	d := &dirServer{existing: map[string]bool{"/photos": true, "/photos/2026": true}}
	srv := httptest.NewServer(d.handler())
	defer srv.Close()

	client := newDirClient(t, srv)

	if errResp := client.CreateDirAll(context.Background(), "/photos/2026/summer"); errResp != nil {
		t.Fatalf("existing parents must not be an error, got: %v", errResp.Error)
	}
	if len(d.created) != 1 || d.created[0] != "/photos/2026/summer" {
		t.Errorf("expected only the missing leaf to be created, got %v", d.created)
	}
}

func TestCreateDirAllIsIdempotent(t *testing.T) {
	d := &dirServer{existing: map[string]bool{"/a": true, "/a/b": true}}
	srv := httptest.NewServer(d.handler())
	defer srv.Close()

	client := newDirClient(t, srv)

	if errResp := client.CreateDirAll(context.Background(), "/a/b"); errResp != nil {
		t.Fatalf("creating an existing directory must succeed, got: %v", errResp.Error)
	}
	if len(d.created) != 0 {
		t.Errorf("nothing should have been created, got %v", d.created)
	}
}

func TestCreateDirAllPathForms(t *testing.T) {
	cases := []struct {
		name string
		path string
		want []string
	}{
		{"absolute", "/a/b", []string{"/a", "/a/b"}},
		{"relative", "a/b", []string{"a", "a/b"}},
		{"disk prefix", "disk:/a/b", []string{"/a", "/a/b"}},
		{"redundant slashes", "/a//b/", []string{"/a", "/a/b"}},
		{"single segment", "/a", []string{"/a"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &dirServer{existing: map[string]bool{}}
			srv := httptest.NewServer(d.handler())
			defer srv.Close()

			if errResp := newDirClient(t, srv).CreateDirAll(context.Background(), tc.path); errResp != nil {
				t.Fatalf("unexpected error: %v", errResp.Error)
			}
			if len(d.created) != len(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, d.created)
			}
			for i, path := range tc.want {
				if d.created[i] != path {
					t.Errorf("request %d: expected %q, got %q", i, path, d.created[i])
				}
			}
		})
	}
}

func TestCreateDirAllPropagatesRealErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "DiskResourceUploadFailedError", Message: "no"})
	}))
	defer srv.Close()

	errResp := newDirClient(t, srv).CreateDirAll(context.Background(), "/a/b")
	if errResp == nil {
		t.Fatal("expected the 403 to be propagated")
	}
	if errResp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403 to be recorded, got %d", errResp.StatusCode)
	}
	if errResp.AlreadyExists() {
		t.Error("a 403 must not be mistaken for an existing directory")
	}
}

func TestCreateDirAllRejectsBadPaths(t *testing.T) {
	client, err := New("token")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", "/a/../b", "/", "a\x00b"} {
		if errResp := client.CreateDirAll(context.Background(), path); errResp == nil {
			t.Errorf("expected %q to be rejected", path)
		}
	}
}

func TestAlreadyExists(t *testing.T) {
	var nilResp *ErrorResponse
	if nilResp.AlreadyExists() {
		t.Error("a nil error is not an existing directory")
	}
	conflict := &ErrorResponse{StatusCode: http.StatusConflict, Error: "DiskPathPointsToExistentDirectoryError"}
	if !conflict.AlreadyExists() {
		t.Error("expected the Yandex existing-directory conflict to be recognised")
	}
	other := &ErrorResponse{StatusCode: http.StatusConflict, Error: "DiskPathDoesntExistsError"}
	if other.AlreadyExists() {
		t.Error("a missing-parent conflict is a real error")
	}
}
