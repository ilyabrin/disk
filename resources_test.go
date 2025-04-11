package disk

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDeleteResourceURL(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name        string
		path        string
		permanently bool
		want        string
	}{
		{
			name:        "Delete temporary file",
			path:        "/path/to/file.txt",
			permanently: false,
			want:        "resources?path=%2Fpath%2Fto%2Ffile.txt&permanent=false",
		},
		{
			name:        "Delete permanent file",
			path:        "/another/path/to/file.jpg",
			permanently: true,
			want:        "resources?path=%2Fanother%2Fpath%2Fto%2Ffile.jpg&permanent=true",
		},
		{
			name:        "Delete file with special characters",
			path:        "/path with spaces/file with &.txt",
			permanently: false,
			want:        "resources?path=%2Fpath+with+spaces%2Ffile+with+%26.txt&permanent=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildDeleteResourceURL(tt.path, tt.permanently)
			assert.Equal(t, tt.want, got)

			// Additional check: ensure the generated URL is valid
			_, err := url.Parse(got)
			assert.Nil(t, err)
		})
	}
}

func TestGetMetadata(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"name": "testdir", "path": "disk:/testdir", "type": "dir"}`))
		}))

	resource, errResp := client.GetMetadata(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, resource)
	assert.Equal(t, "testdir", resource.Name)
	assert.Equal(t, "disk:/testdir", resource.Path)
}

func TestUpdateMetadata(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(
				`{
					"name": "testdir",
					"custom_properties": {
						"key_01": "value_01",
						"key_02": "value_02"
					}
				}`))
		}))

	newMeta := map[string]map[string]string{"custom_properties": {
		"key_01": "value_01",
		"key_02": "value_02",
	}}

	resource, errResp := client.UpdateMetadata(context.Background(), "testdir", newMeta)
	assert.Nil(t, errResp)
	assert.NotNil(t, resource)
	assert.Equal(t, "testdir", resource.Name)

	// Fix: Directly assert and access the CustomProperties field
	customProperties := resource.CustomProperties
	assert.NotNil(t, customProperties)
	assert.IsType(t, map[string]string{}, customProperties)
	assert.Len(t, customProperties, 2)
	// Check the values of the custom properties
	assert.Equal(t, "value_01", customProperties["key_01"])
	assert.Equal(t, "value_02", customProperties["key_02"])
}

func TestCreateDir(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources?path=disk:/testdir", "method": "PUT", "templated": false}`))
		}))

	link, errResp := client.CreateDir(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "PUT", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources?path=disk:/testdir", link.Href)
}

func TestCopyResource(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources/copy", "method": "POST", "templated": false}`))
		}))

	link, errResp := client.CopyResource(context.Background(), "source", "destination")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "POST", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources/copy", link.Href)
}

func TestGetDownloadURL(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://downloader.disk.yandex.net/disk/resource", "method": "GET", "templated": false}`))
		}))

	link, errResp := client.GetDownloadURL(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "GET", link.Method)
	assert.Equal(t, "https://downloader.disk.yandex.net/disk/resource", link.Href)
}

func TestGetSortedFiles(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"items": [{"name": "file1.txt", "path": "disk:/file1.txt"}]}`))
		}))

	files, errResp := client.GetSortedFiles(context.Background())
	assert.Nil(t, errResp)
	assert.NotNil(t, files)
	assert.Len(t, files.Items, 1)
	assert.Equal(t, "file1.txt", files.Items[0].Name)
	assert.Equal(t, "disk:/file1.txt", files.Items[0].Path)
}

func TestMoveResource(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources/move", "method": "POST", "templated": false}`))
		}))

	link, errResp := client.MoveResource(context.Background(), "source", "destination")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "POST", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources/move", link.Href)
}

func TestGetLastUploadedResources(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"items": [{"name": "file1.txt", "path": "disk:/file1.txt"}]}`))
		}))

	resources, errResp := client.GetLastUploadedResources(context.Background())
	assert.Nil(t, errResp)
	assert.NotNil(t, resources)
	assert.Len(t, resources.Items, 1)
	assert.Equal(t, "file1.txt", resources.Items[0].Name)
	assert.Equal(t, "disk:/file1.txt", resources.Items[0].Path)
}

func TestGetPublicResources(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"items": [{"name": "file1.txt", "path": "disk:/file1.txt"}]}`))
		}))

	resources, errResp := client.GetPublicResources(context.Background())
	assert.Nil(t, errResp)
	assert.NotNil(t, resources)
	assert.Len(t, resources.Items, 1)
	assert.Equal(t, "file1.txt", resources.Items[0].Name)
	assert.Equal(t, "disk:/file1.txt", resources.Items[0].Path)
}

func TestPublishResource(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources/publish", "method": "POST", "templated": false}`))
		}))

	link, errResp := client.PublishResource(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "POST", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources/publish", link.Href)
}

func TestUnpublishResource(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources/unpublish", "method": "POST", "templated": false}`))
		}))

	link, errResp := client.UnpublishResource(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "POST", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources/unpublish", link.Href)
}

func TestGetLinkForUpload(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources/upload", "method": "POST", "templated": false}`))
		}))

	link, errResp := client.GetLinkForUpload(context.Background(), "testdir")
	assert.Nil(t, errResp)
	assert.NotNil(t, link)
	assert.Equal(t, "POST", link.Method)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources/upload", link.Href)
}

func TestUploadFile(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
			w.Write([]byte(`{"href": "https://uploader.disk.yandex.net/upload", "method": "PUT", "templated": false}`))
		}))

	resp, errResp := client.UploadFile(context.Background(), "testdir/testfile", "file content")
	assert.Nil(t, errResp)
	assert.IsType(t, &Link{}, resp)
}

func TestDeleteResource_EmptyPath(t *testing.T) {
	client := &Client{} // Initialize the client with any necessary fields

	ctx := context.Background()
	path := ""
	permanently := true

	err := client.DeleteResource(ctx, path, permanently)

	if err == nil {
		t.Errorf("expected an error when path is empty, but got nil")
	}

	expectedError := "delete error: empty path"
	if err.Error() != expectedError {
		t.Errorf("expected error: %s, but got: %s", expectedError, err.Error())
	}
}

// func TestDeleteResource(t *testing.T) {
// 	t.Run("Successful deletion", func(t *testing.T) {
// 		client := mockedHttpClient(
// 			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				assert.NotEmpty(t, r.Header.Get("Authorization"))
// 				assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
// 				w.WriteHeader(http.StatusNoContent) // Simulate 204 No Content
// 			}))

// 		err := client.DeleteResource(context.Background(), "testdir", true)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Empty path error", func(t *testing.T) {
// 		client := mockedHttpClient(nil)

// 		err := client.DeleteResource(context.Background(), "", true)
// 		assert.Error(t, err)
// 		assert.Equal(t, "delete error: empty path", err.Error())
// 	})

// 	t.Run("API error response", func(t *testing.T) {
// 		client := mockedHttpClient(
// 			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				assert.NotEmpty(t, r.Header.Get("Authorization"))
// 				assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
// 				w.WriteHeader(http.StatusBadRequest)
// 				w.Write([]byte(`{"error": "invalid_path", "description": "The specified path is invalid."}`))
// 			}))

// 		err := client.DeleteResource(context.Background(), "*$invalid/path", true)
// 		// assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "delete failed")
// 		assert.Contains(t, err.Error(), "invalid_path")
// 	})

// 	t.Run("Network error", func(t *testing.T) {
// 		client := mockedHttpClient(nil)

// 		// Simulate a network error by using a nil HTTP handler
// 		err := client.DeleteResource(context.Background(), "testdir", true)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "delete failed")
// 	})
// }
