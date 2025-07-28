package disk

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestTrashOperations(t *testing.T) {
	t.Run("RestoreFromTrash with basic parameters", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, "trash/resources/restore") {
				t.Errorf("Expected trash restore endpoint, got %s", r.URL.Path)
			}

			// Check query parameters
			path := r.URL.Query().Get("path")
			if path != "/test/file.txt" {
				t.Errorf("Expected path '/test/file.txt', got '%s'", path)
			}

			w.WriteHeader(200)
			w.Write([]byte(`{"href": "https://example.com", "method": "GET"}`))
		})

		link, err := client.RestoreFromTrash(context.Background(), "/test/file.txt", false, "")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if link == nil {
			t.Fatal("Expected link to be returned")
		}
	})

	t.Run("RestoreFromTrash with overwrite and name", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			// Check query parameters
			if r.URL.Query().Get("overwrite") != "true" {
				t.Error("Expected overwrite=true")
			}
			if r.URL.Query().Get("name") != "new_name.txt" {
				t.Error("Expected name=new_name.txt")
			}

			w.WriteHeader(202)
			w.Write([]byte(`{"href": "https://example.com/operation/123", "method": "GET"}`))
		})

		link, err := client.RestoreFromTrash(context.Background(), "/test/file.txt", true, "new_name.txt")
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if link.Href != "https://example.com/operation/123" {
			t.Errorf("Expected operation link, got: %s", link.Href)
		}
	})

	t.Run("RestoreFromTrash with empty path", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not make request with empty path")
		})

		_, err := client.RestoreFromTrash(context.Background(), "", false, "")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}

func TestListTrashResources(t *testing.T) {
	t.Run("ListTrashResources with basic parameters", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, "trash/resources") {
				t.Errorf("Expected trash resources endpoint, got %s", r.URL.Path)
			}

			w.WriteHeader(200)
			w.Write([]byte(`{
				"items": [
					{
						"path": "/trash/deleted_file.txt",
						"name": "deleted_file.txt",
						"origin_path": "/original/deleted_file.txt",
						"deleted": "2023-01-01T10:00:00Z",
						"type": "file",
						"size": 1024
					}
				],
				"limit": 20,
				"offset": 0
			}`))
		})

		trashList, err := client.ListTrashResources(context.Background(), "", 0, 0)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if len(trashList.Items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(trashList.Items))
		}

		item := trashList.Items[0]
		if item.Name != "deleted_file.txt" {
			t.Errorf("Expected name 'deleted_file.txt', got '%s'", item.Name)
		}
		if item.OriginPath != "/original/deleted_file.txt" {
			t.Errorf("Expected origin path '/original/deleted_file.txt', got '%s'", item.OriginPath)
		}
	})

	t.Run("ListTrashResources with pagination", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			limit := r.URL.Query().Get("limit")
			offset := r.URL.Query().Get("offset")

			if limit != "10" {
				t.Errorf("Expected limit=10, got %s", limit)
			}
			if offset != "20" {
				t.Errorf("Expected offset=20, got %s", offset)
			}

			w.WriteHeader(200)
			w.Write([]byte(`{"items": [], "limit": 10, "offset": 20}`))
		})

		trashList, err := client.ListTrashResources(context.Background(), "", 10, 20)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if trashList.Limit != 10 {
			t.Errorf("Expected limit 10, got %d", trashList.Limit)
		}
		if trashList.Offset != 20 {
			t.Errorf("Expected offset 20, got %d", trashList.Offset)
		}
	})
}

func TestEmptyTrash(t *testing.T) {
	t.Run("EmptyTrash successfully", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			if !strings.Contains(r.URL.Path, "trash/resources") {
				t.Errorf("Expected trash resources endpoint, got %s", r.URL.Path)
			}

			w.WriteHeader(204) // No content
		})

		err := client.EmptyTrash(context.Background(), "", false)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
	})

	t.Run("EmptyTrash with specific path", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get("path")
			if path != "/trash/folder" {
				t.Errorf("Expected path '/trash/folder', got '%s'", path)
			}

			w.WriteHeader(202) // Async operation
		})

		err := client.EmptyTrash(context.Background(), "/trash/folder", false)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
	})

	t.Run("EmptyTrash with force", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			forceAsync := r.URL.Query().Get("force_async")
			if forceAsync != "false" {
				t.Errorf("Expected force_async=false, got '%s'", forceAsync)
			}

			w.WriteHeader(200)
		})

		err := client.EmptyTrash(context.Background(), "", true)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}
	})
}

func TestGetTrashResourceMetadata(t *testing.T) {
	t.Run("GetTrashResourceMetadata successfully", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}

			path := r.URL.Query().Get("path")
			if path != "/trash/test_file.txt" {
				t.Errorf("Expected path '/trash/test_file.txt', got '%s'", path)
			}

			w.WriteHeader(200)
			w.Write([]byte(`{
				"path": "/trash/test_file.txt",
				"name": "test_file.txt",
				"origin_path": "/original/test_file.txt",
				"deleted": "2023-01-01T10:00:00Z",
				"type": "file",
				"size": 2048,
				"md5": "abcdef123456"
			}`))
		})

		metadata, err := client.GetTrashResourceMetadata(context.Background(), "/trash/test_file.txt", nil)
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if metadata.Name != "test_file.txt" {
			t.Errorf("Expected name 'test_file.txt', got '%s'", metadata.Name)
		}
		if metadata.OriginPath != "/original/test_file.txt" {
			t.Errorf("Expected origin path '/original/test_file.txt', got '%s'", metadata.OriginPath)
		}
		if metadata.Size != 2048 {
			t.Errorf("Expected size 2048, got %d", metadata.Size)
		}
	})

	t.Run("GetTrashResourceMetadata with fields", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			fields := r.URL.Query()["fields"]
			expectedFields := []string{"name", "size", "md5"}

			if len(fields) != len(expectedFields) {
				t.Errorf("Expected %d fields, got %d", len(expectedFields), len(fields))
			}

			for i, field := range fields {
				if i < len(expectedFields) && field != expectedFields[i] {
					t.Errorf("Expected field '%s', got '%s'", expectedFields[i], field)
				}
			}

			w.WriteHeader(200)
			w.Write([]byte(`{"name": "test_file.txt", "size": 2048, "md5": "abcdef123456"}`))
		})

		metadata, err := client.GetTrashResourceMetadata(context.Background(), "/trash/test_file.txt", []string{"name", "size", "md5"})
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if metadata.Name != "test_file.txt" {
			t.Errorf("Expected name 'test_file.txt', got '%s'", metadata.Name)
		}
	})

	t.Run("GetTrashResourceMetadata with empty path", func(t *testing.T) {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Should not make request with empty path")
		})

		_, err := client.GetTrashResourceMetadata(context.Background(), "", nil)
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}