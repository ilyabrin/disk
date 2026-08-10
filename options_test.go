package disk

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestValidatePath(t *testing.T) {
	valid := []string{
		"/",
		"disk:/foo/bar.txt",
		"/reports/report..final.pdf",
		"/a..b/c",
	}
	for _, p := range valid {
		if err := validatePath(p); err != nil {
			t.Errorf("validatePath(%q) = %v, want nil", p, err)
		}
	}

	invalid := []string{
		"",
		"/foo/../../etc/passwd",
		"../secrets",
		"/foo\x00bar",
	}
	for _, p := range invalid {
		if err := validatePath(p); err == nil {
			t.Errorf("validatePath(%q) = nil, want an error", p)
		}
	}
}

// capturedQuery runs fn against a stub server and returns the query the client sent.
func capturedQuery(t *testing.T, body string, fn func(c *Client)) url.Values {
	t.Helper()

	var query url.Values
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
	fn(client)
	return query
}

func TestResourceOptionsQuery(t *testing.T) {
	query := capturedQuery(t, `{"name":"docs"}`, func(c *Client) {
		c.GetMetadataWithOptions(context.Background(), "/docs", &ResourceOptions{
			Limit:       50,
			Offset:      10,
			Sort:        "-modified",
			PreviewSize: "M",
			PreviewCrop: true,
			Fields:      []string{"name", "_embedded.items.path"},
		})
	})

	want := map[string]string{
		"path":         "/docs",
		"limit":        "50",
		"offset":       "10",
		"sort":         "-modified",
		"preview_size": "M",
		"preview_crop": "true",
		"fields":       "name,_embedded.items.path",
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("query[%s] = %q, want %q", key, got, value)
		}
	}
}

func TestFilesOptionsQuery(t *testing.T) {
	query := capturedQuery(t, `{"items":[]}`, func(c *Client) {
		c.GetSortedFilesWithOptions(context.Background(),
			&PaginationOptions{Limit: 5},
			&FilesOptions{
				MediaType:   []string{"image", "video"},
				Sort:        "size",
				PreviewSize: "120x240",
				Fields:      []string{"items.name"},
			})
	})

	if got := query.Get("media_type"); got != "image,video" {
		t.Errorf("media_type = %q, want %q", got, "image,video")
	}
	if got := query.Get("sort"); got != "size" {
		t.Errorf("sort = %q, want %q", got, "size")
	}
	if got := query.Get("preview_size"); got != "120x240" {
		t.Errorf("preview_size = %q, want %q", got, "120x240")
	}
	if got := query.Get("fields"); got != "items.name" {
		t.Errorf("fields = %q, want %q", got, "items.name")
	}
}

func TestPublicResourceOptionsQuery(t *testing.T) {
	query := capturedQuery(t, `{"name":"shared"}`, func(c *Client) {
		c.GetMetadataForPublicResourceWithOptions(context.Background(), "key123", &PublicResourceOptions{
			Path:        "/nested/file.txt",
			Sort:        "name",
			Limit:       30,
			Offset:      5,
			PreviewSize: "L",
			PreviewCrop: true,
		})
	})

	want := map[string]string{
		"public_key":   "key123",
		"path":         "/nested/file.txt",
		"sort":         "name",
		"limit":        "30",
		"offset":       "5",
		"preview_size": "L",
		"preview_crop": "true",
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("query[%s] = %q, want %q", key, got, value)
		}
	}
}

func TestPublicDownloadAndSaveOptions(t *testing.T) {
	t.Run("download link for a nested path", func(t *testing.T) {
		query := capturedQuery(t, `{"href":"https://example.test/file"}`, func(c *Client) {
			c.GetDownloadURLForPublicResourceAt(context.Background(), "key123", "/nested/file.txt")
		})
		if got := query.Get("path"); got != "/nested/file.txt" {
			t.Errorf("path = %q, want %q", got, "/nested/file.txt")
		}
	})

	t.Run("save-to-disk with name", func(t *testing.T) {
		query := capturedQuery(t, `{"href":"https://example.test/op"}`, func(c *Client) {
			c.SavePublicResourceWithOptions(context.Background(), "key123", &SavePublicResourceOptions{
				Path: "/nested/file.txt",
				Name: "copy.txt",
			})
		})
		if got := query.Get("name"); got != "copy.txt" {
			t.Errorf("name = %q, want %q", got, "copy.txt")
		}
		if got := query.Get("path"); got != "/nested/file.txt" {
			t.Errorf("path = %q, want %q", got, "/nested/file.txt")
		}
	})
}
