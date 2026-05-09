package disk

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetadataWithOptions(t *testing.T) {
	t.Run("nil opts falls through to GetMetadata", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
				// no extra query params expected
				assert.Empty(t, r.URL.Query().Get("limit"))
				assert.Empty(t, r.URL.Query().Get("sort"))
				w.Write([]byte(`{"name":"testdir","type":"dir","path":"disk:/testdir"}`))
			}))

		resource, errResp := client.GetMetadataWithOptions(context.Background(), "testdir", nil)
		assert.Nil(t, errResp)
		assert.IsType(t, &Resource{}, resource)
		assert.Equal(t, "testdir", resource.Name)
	})

	t.Run("opts with limit, offset, sort are sent as query params", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				assert.Equal(t, "10", q.Get("limit"))
				assert.Equal(t, "5", q.Get("offset"))
				assert.Equal(t, "-name", q.Get("sort"))
				w.Write([]byte(`{"name":"dir","type":"dir","path":"disk:/dir"}`))
			}))

		opts := &ResourceOptions{Limit: 10, Offset: 5, Sort: "-name"}
		resource, errResp := client.GetMetadataWithOptions(context.Background(), "dir", opts)
		assert.Nil(t, errResp)
		assert.NotNil(t, resource)
	})

	t.Run("opts with preview_size and preview_crop", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				assert.Equal(t, "M", q.Get("preview_size"))
				assert.Equal(t, "true", q.Get("preview_crop"))
				w.Write([]byte(`{"name":"photo.jpg","type":"file","path":"disk:/photo.jpg"}`))
			}))

		opts := &ResourceOptions{PreviewSize: "M", PreviewCrop: true}
		resource, errResp := client.GetMetadataWithOptions(context.Background(), "photo.jpg", opts)
		assert.Nil(t, errResp)
		assert.NotNil(t, resource)
	})

	t.Run("zero Limit and Offset are not sent", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				assert.Empty(t, q.Get("limit"))
				assert.Empty(t, q.Get("offset"))
				w.Write([]byte(`{"name":"dir","type":"dir","path":"disk:/dir"}`))
			}))

		opts := &ResourceOptions{Limit: 0, Offset: 0}
		resource, errResp := client.GetMetadataWithOptions(context.Background(), "dir", opts)
		assert.Nil(t, errResp)
		assert.NotNil(t, resource)
	})

	t.Run("empty path returns error", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		resource, errResp := client.GetMetadataWithOptions(context.Background(), "", nil)
		assert.Nil(t, resource)
		assert.NotNil(t, errResp)
	})

	t.Run("API error response is returned", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"DiskNotFoundError","description":"not found"}`))
			}))

		resource, errResp := client.GetMetadataWithOptions(context.Background(), "missing", nil)
		assert.Nil(t, resource)
		assert.NotNil(t, errResp)
		assert.Equal(t, "DiskNotFoundError", errResp.Error)
	})
}

func TestDownloadFileToPath(t *testing.T) {
	// Build a client whose mock serves both the download-URL endpoint and the
	// actual file download. The testTransport routes all requests to the same
	// httptest server, so the handler has to distinguish by path.
	t.Run("successful download", func(t *testing.T) {
		fileContent := "hello disk"

		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "resources/download") {
					// Return the download link pointing back to our test server
					addr := r.Host
					downloadURL := "http://" + addr + "/file-content"
					w.Write([]byte(`{"href":"` + downloadURL + `","method":"GET","templated":false}`))
					return
				}
				if r.URL.Path == "/file-content" {
					w.Write([]byte(fileContent))
					return
				}
				http.NotFound(w, r)
			}))

		dest := filepath.Join(t.TempDir(), "out.txt")
		err := client.DownloadFileToPath(context.Background(), "/test.txt", dest, nil)
		assert.NoError(t, err)

		data, readErr := os.ReadFile(dest)
		assert.NoError(t, readErr)
		assert.Equal(t, fileContent, string(data))
	})

	t.Run("fails when remote path is empty", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		err := client.DownloadFileToPath(context.Background(), "", "/tmp/out.txt", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remote path")
	})

	t.Run("fails when local path is empty", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		err := client.DownloadFileToPath(context.Background(), "/test.txt", "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "local path")
	})

	t.Run("fails when local file exists and Overwrite is false", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		dest := filepath.Join(t.TempDir(), "existing.txt")
		os.WriteFile(dest, []byte("old"), 0644)

		err := client.DownloadFileToPath(context.Background(), "/test.txt", dest, &DownloadOptions{Overwrite: false})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("overwrites existing file when Overwrite is true", func(t *testing.T) {
		newContent := "new content"

		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "resources/download") {
					addr := r.Host
					downloadURL := "http://" + addr + "/file-content"
					w.Write([]byte(`{"href":"` + downloadURL + `","method":"GET","templated":false}`))
					return
				}
				if r.URL.Path == "/file-content" {
					w.Write([]byte(newContent))
					return
				}
				http.NotFound(w, r)
			}))

		dest := filepath.Join(t.TempDir(), "existing.txt")
		os.WriteFile(dest, []byte("old content"), 0644)

		err := client.DownloadFileToPath(context.Background(), "/test.txt", dest, &DownloadOptions{Overwrite: true})
		assert.NoError(t, err)

		data, _ := os.ReadFile(dest)
		assert.Equal(t, newContent, string(data))
	})

	t.Run("progress callback is invoked", func(t *testing.T) {
		fileContent := "progress test content"
		var callbackInvoked bool

		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "resources/download") {
					addr := r.Host
					downloadURL := "http://" + addr + "/file-content"
					w.Write([]byte(`{"href":"` + downloadURL + `","method":"GET","templated":false}`))
					return
				}
				if r.URL.Path == "/file-content" {
					w.Write([]byte(fileContent))
					return
				}
				http.NotFound(w, r)
			}))

		dest := filepath.Join(t.TempDir(), "progress.txt")
		opts := &DownloadOptions{
			Progress: func(p DownloadProgress) {
				callbackInvoked = true
				assert.GreaterOrEqual(t, p.BytesDownloaded, int64(0))
			},
		}
		err := client.DownloadFileToPath(context.Background(), "/test.txt", dest, opts)
		assert.NoError(t, err)
		assert.True(t, callbackInvoked)
	})

	t.Run("fails when download URL API returns error", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"DiskNotFoundError","description":"not found"}`))
			}))

		dest := filepath.Join(t.TempDir(), "out.txt")
		err := client.DownloadFileToPath(context.Background(), "/missing.txt", dest, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "download URL")
	})
}

func TestDownloadFileToPathWithProgress(t *testing.T) {
	fileContent := "wrapper test"
	var lastProgress DownloadProgress

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "resources/download") {
				addr := r.Host
				// Build the full URL manually without using url.Parse on a relative path
				downloadHref := (&url.URL{Scheme: "http", Host: addr, Path: "/file-content"}).String()
				w.Write([]byte(`{"href":"` + downloadHref + `","method":"GET","templated":false}`))
				return
			}
			if r.URL.Path == "/file-content" {
				w.Write([]byte(fileContent))
				return
			}
			http.NotFound(w, r)
		}))

	dest := filepath.Join(t.TempDir(), "wrapper.txt")
	err := client.DownloadFileToPathWithProgress(context.Background(), "/test.txt", dest, false,
		func(p DownloadProgress) { lastProgress = p })

	assert.NoError(t, err)
	assert.Greater(t, lastProgress.BytesDownloaded, int64(0))

	data, _ := os.ReadFile(dest)
	assert.Equal(t, fileContent, string(data))
}
