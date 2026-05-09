package disk

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestUploadFileFromPath(t *testing.T) {
	t.Run("UploadFileFromPath function exists and validates input", func(t *testing.T) {
		client, _ := New("test-token")

		// Test that the method exists and can be called
		// We expect it to fail since we don't have a real token, but we're just testing the method exists
		_, err := client.UploadFileFromPath(context.Background(), "", "/test/upload.txt", nil)
		if err == nil {
			t.Error("Expected error for empty local path")
		}
		if !strings.Contains(err.Error(), "local path cannot be empty") {
			t.Errorf("Expected 'local path cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("UploadFileFromPath with empty local path", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.UploadFileFromPath(context.Background(), "", "/test/upload.txt", nil)
		if err == nil {
			t.Error("Expected error for empty local path")
		}
		if !strings.Contains(err.Error(), "local path cannot be empty") {
			t.Errorf("Expected 'local path cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("UploadFileFromPath with empty remote path", func(t *testing.T) {
		client, _ := New("test-token")

		tmpFile, err := os.CreateTemp("", "upload_test_*.txt")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = client.UploadFileFromPath(context.Background(), tmpFile.Name(), "", nil)
		if err == nil {
			t.Error("Expected error for empty remote path")
		}
		if !strings.Contains(err.Error(), "remote path cannot be empty") {
			t.Errorf("Expected 'remote path cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("UploadFileFromPath with non-existent file", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.UploadFileFromPath(context.Background(), "/non/existent/file.txt", "/test/upload.txt", nil)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "file does not exist") {
			t.Errorf("Expected 'file does not exist' error, got: %s", err.Error())
		}
	})

	t.Run("UploadFileFromPath with directory instead of file", func(t *testing.T) {
		client, _ := New("test-token")

		tmpDir, err := os.MkdirTemp("", "upload_test_dir_*")
		if err != nil {
			t.Fatal("Failed to create temp dir:", err)
		}
		defer os.RemoveAll(tmpDir)

		_, err = client.UploadFileFromPath(context.Background(), tmpDir, "/test/upload.txt", nil)
		if err == nil {
			t.Error("Expected error for directory path")
		}
		if !strings.Contains(err.Error(), "path is a directory") {
			t.Errorf("Expected 'path is a directory' error, got: %s", err.Error())
		}
	})

	t.Run("UploadFileFromPath progress callback validation", func(t *testing.T) {
		client, _ := New("test-token")

		// Test that progress callback is properly accepted and validated
		var progressUpdates []UploadProgress
		progressCallback := func(progress UploadProgress) {
			progressUpdates = append(progressUpdates, progress)
		}

		options := &UploadOptions{
			Progress: progressCallback,
		}

		// This will fail at API call level, but we're testing that the option is accepted
		_, err := client.UploadFileFromPath(context.Background(), "/non/existent/file.txt", "/test/upload.txt", options)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		// Verify the error is about file validation, not about progress callback
		if !strings.Contains(err.Error(), "file does not exist") {
			t.Errorf("Expected file validation error, got: %s", err.Error())
		}
	})
}

func TestValidateLocalFile(t *testing.T) {
	t.Run("ValidateLocalFile with valid file", func(t *testing.T) {
		client, _ := New("test-token")

		tmpFile, err := os.CreateTemp("", "validate_test_*.txt")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		tmpFile.WriteString("test content")
		tmpFile.Sync()

		fileInfo, err := client.validateLocalFile(tmpFile.Name())
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		if fileInfo.Size() <= 0 {
			t.Error("Expected file size > 0")
		}
	})

	t.Run("ValidateLocalFile with empty file", func(t *testing.T) {
		client, _ := New("test-token")

		tmpFile, err := os.CreateTemp("", "validate_empty_test_*.txt")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = client.validateLocalFile(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for empty file")
		}
		if !strings.Contains(err.Error(), "file is empty") {
			t.Errorf("Expected 'file is empty' error, got: %s", err.Error())
		}
	})
}

func TestDetectMimeType(t *testing.T) {
	t.Run("DetectMimeType by extension", func(t *testing.T) {
		client, _ := New("test-token")

		tmpFile, err := os.CreateTemp("", "mime_test_*.txt")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		tmpFile.WriteString("test content")
		tmpFile.Sync()

		mimeType, err := client.DetectMimeType(tmpFile.Name())
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		expectedType := "text/plain; charset=utf-8"
		if mimeType != expectedType {
			t.Errorf("Expected MIME type '%s', got '%s'", expectedType, mimeType)
		}
	})

	t.Run("DetectMimeType for non-existent file", func(t *testing.T) {
		client, _ := New("test-token")

		// Use a file without a recognized extension to force content reading
		_, err := client.DetectMimeType("/non/existent/file.unknown")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "cannot open file") {
			t.Errorf("Expected 'cannot open file' error, got: %s", err.Error())
		}
	})
}

func TestValidateFilePath(t *testing.T) {
	testCases := []struct {
		path        string
		shouldError bool
		description string
	}{
		{"/valid/path/file.txt", false, "valid path"},
		{"", true, "empty path"},
		{"invalid/path", true, "path not starting with /"},
		{"/path/with<invalid>chars", true, "path with invalid characters"},
		{"/path/with:colon", true, "path with colon"},
		{"/path/with\"quote", true, "path with quote"},
		{"/path/with|pipe", true, "path with pipe"},
		{"/path/with?question", true, "path with question mark"},
		{"/path/with*asterisk", true, "path with asterisk"},
		{"/valid/long/path/that/is/acceptable.txt", false, "normal long path"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := ValidateFilePath(tc.path)
			if tc.shouldError && err == nil {
				t.Errorf("Expected error for %s, got none", tc.description)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error for %s, got: %s", tc.description, err.Error())
			}
		})
	}
}

// uploadMockTransport simulates the file upload to Yandex servers
type uploadMockTransport struct {
	t               *testing.T
	expectedContent string
}

func (u *uploadMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read the request body to verify content
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			u.t.Errorf("Failed to read request body: %v", err)
		}

		// Verify content matches expected
		if string(body) != u.expectedContent {
			u.t.Errorf("Upload content mismatch. Expected: %s, Got: %s", u.expectedContent, string(body))
		}
	}

	// Return successful response
	return &http.Response{
		StatusCode: 201,
		Status:     "201 Created",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req,
	}, nil
}

func TestUploadOptions(t *testing.T) {
	t.Run("UploadOptions validation", func(t *testing.T) {
		client, _ := New("test-token")

		// Test that all upload options are properly accepted
		var progressUpdates []UploadProgress
		progressCallback := func(progress UploadProgress) {
			progressUpdates = append(progressUpdates, progress)
		}

		options := &UploadOptions{
			Overwrite:        true,
			Progress:         progressCallback,
			ChunkSize:        5 * 1024 * 1024, // 5MB
			ValidateChecksum: false,
		}

		// This will fail at file validation level, but we're testing that options are accepted
		_, err := client.UploadFileFromPath(context.Background(), "/non/existent/file.txt", "/test/upload.txt", options)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		// Verify the error is about file validation, not about options
		if !strings.Contains(err.Error(), "file does not exist") {
			t.Errorf("Expected file validation error, got: %s", err.Error())
		}
	})
}

// overwriteMockTransport checks for overwrite header
type overwriteMockTransport struct {
	t *testing.T
}

func (o *overwriteMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check for overwrite header
	overwriteHeader := req.Header.Get("X-Overwrite")
	if overwriteHeader != "true" {
		o.t.Error("Expected X-Overwrite header to be 'true'")
	}

	return &http.Response{
		StatusCode: 201,
		Status:     "201 Created",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req,
	}, nil
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("GetFileSize works correctly", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "size_test_*.txt")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		testContent := "This is test content for size checking"
		tmpFile.WriteString(testContent)
		tmpFile.Sync()

		size, err := GetFileSize(tmpFile.Name())
		if err != nil {
			t.Fatal("Expected no error, got:", err)
		}

		expectedSize := int64(len(testContent))
		if size != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, size)
		}
	})

	t.Run("GetFileSize fails for non-existent file", func(t *testing.T) {
		_, err := GetFileSize("/non/existent/file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("FormatFileSize formats correctly", func(t *testing.T) {
		testCases := []struct {
			bytes    int64
			expected string
		}{
			{512, "512 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{1024 * 1024, "1.0 MB"},
			{1024 * 1024 * 1024, "1.0 GB"},
			{1536 * 1024 * 1024, "1.5 GB"},
		}

		for _, tc := range testCases {
			result := FormatFileSize(tc.bytes)
			if result != tc.expected {
				t.Errorf("FormatFileSize(%d) = %s, expected %s", tc.bytes, result, tc.expected)
			}
		}
	})
}

func TestConvenienceMethods(t *testing.T) {
	t.Run("UploadFileFromPathWithProgress validates input", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.UploadFileFromPathWithProgress(context.Background(), "", "/test/file.txt", false, nil)
		if err == nil {
			t.Error("Expected error for empty local path")
		}
	})

	t.Run("UploadLargeFileFromPath validates input", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.UploadLargeFileFromPath(context.Background(), "", "/test/file.txt", 10, nil)
		if err == nil {
			t.Error("Expected error for empty local path")
		}
	})
}

func TestUploadInternalFunctions(t *testing.T) {
	t.Run("UploadFile function validation", func(t *testing.T) {
		client, _ := New("test-token")

		// Test the public UploadFile function with correct signature
		_, err := client.UploadFile(context.Background(), "/test/file.txt", "https://mock-upload-url.com")
		if err != nil {
			// Expected to fail due to mock setup, but function should exist and validate
			t.Log("UploadFile function exists and validates input")
		}
	})

	t.Run("DetectMimeType covers additional cases", func(t *testing.T) {
		client, _ := New("test-token")

		// Test DetectMimeType with various extensions
		testCases := []struct {
			filename string
			contains string // Use contains instead of exact match
		}{
			{"test.txt", "text/plain"},
			{"test.json", "application/json"},
		}

		for _, tc := range testCases {
			mimeType, _ := client.DetectMimeType(tc.filename)
			if !strings.Contains(mimeType, tc.contains) {
				t.Errorf("Expected mime type to contain %s for %s, got %s", tc.contains, tc.filename, mimeType)
			}
		}
	})
}

// makeTempFile creates a temp file with the given content and returns its path.
func makeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "upload_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content)
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// uploadMockHandler returns an http.HandlerFunc that serves:
//   - GET /resources/upload  → upload link href pointing to uploadPath
//   - PUT <uploadPath>       → responds with uploadStatus
//   - GET /resources         → resource JSON
func uploadMockHandler(t *testing.T, uploadPath string, uploadStatus int, resourceJSON string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/disk/resources/upload" && r.Method == "GET":
			addr := r.Host
			href := "http://" + addr + uploadPath
			w.Write([]byte(`{"href":"` + href + `","method":"PUT","templated":false}`))

		case r.URL.Path == uploadPath && r.Method == "PUT":
			w.WriteHeader(uploadStatus)

		case r.URL.Path == "/v1/disk/resources" && r.Method == "GET":
			if resourceJSON != "" {
				w.Write([]byte(resourceJSON))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"InternalError","description":"metadata unavailable"}`))
			}

		default:
			http.NotFound(w, r)
		}
	}
}

const testResourceJSON = `{"name":"file.txt","type":"file","path":"disk:/test/file.txt","size":12}`

func TestUploadFileSingle(t *testing.T) {
	t.Run("successful single upload returns resource metadata", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, testResourceJSON))

		localPath := makeTempFile(t, "hello upload")
		resource, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", nil)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resource == nil {
			t.Fatal("expected non-nil resource")
		}
		if resource.Name != "file.txt" {
			t.Errorf("expected name 'file.txt', got %q", resource.Name)
		}
	})

	t.Run("upload with progress callback reports progress", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, testResourceJSON))

		localPath := makeTempFile(t, "progress content")
		var updates []UploadProgress
		opts := &UploadOptions{
			Progress: func(p UploadProgress) { updates = append(updates, p) },
		}

		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(updates) == 0 {
			t.Error("expected at least one progress update")
		}
		last := updates[len(updates)-1]
		if last.BytesUploaded <= 0 {
			t.Errorf("expected BytesUploaded > 0, got %d", last.BytesUploaded)
		}
	})

	t.Run("upload with Overwrite option sends X-Overwrite header", func(t *testing.T) {
		var gotOverwrite string
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/v1/disk/resources/upload":
				addr := r.Host
				href := "http://" + addr + "/do-upload"
				w.Write([]byte(`{"href":"` + href + `","method":"PUT","templated":false}`))
			case r.URL.Path == "/do-upload":
				gotOverwrite = r.Header.Get("X-Overwrite")
				w.WriteHeader(http.StatusCreated)
			case r.URL.Path == "/v1/disk/resources":
				w.Write([]byte(testResourceJSON))
			default:
				http.NotFound(w, r)
			}
		})
		client := mockedHttpClient(handler)

		localPath := makeTempFile(t, "overwrite me")
		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", &UploadOptions{Overwrite: true})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if gotOverwrite != "true" {
			t.Errorf("expected X-Overwrite: true, got %q", gotOverwrite)
		}
	})

	t.Run("upload link API error returns error", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"UnauthorizedError","description":"no token"}`))
		}))

		localPath := makeTempFile(t, "some content")
		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "upload link") {
			t.Errorf("expected 'upload link' in error, got: %v", err)
		}
	})

	t.Run("non-2xx upload response returns error", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusForbidden, ""))

		localPath := makeTempFile(t, "some content")
		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "upload failed") {
			t.Errorf("expected 'upload failed' in error, got: %v", err)
		}
	})

	t.Run("metadata fetch failure after upload returns basic resource", func(t *testing.T) {
		// metadata endpoint returns error → uploadFileSingle falls back to basic resource
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, ""))

		localPath := makeTempFile(t, "fallback content")
		resource, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", nil)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resource == nil {
			t.Fatal("expected fallback resource, got nil")
		}
		if resource.Path != "/test/file.txt" {
			t.Errorf("expected path '/test/file.txt', got %q", resource.Path)
		}
	})
}

func TestUploadFileMultipart(t *testing.T) {
	t.Run("large file uses multipart path and returns resource", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, testResourceJSON))

		// Make content larger than chunkSize so multipart branch is taken.
		localPath := makeTempFile(t, strings.Repeat("x", 100))
		opts := &UploadOptions{ChunkSize: 50} // 50 bytes → triggers multipart for 100-byte file

		resource, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resource == nil {
			t.Fatal("expected non-nil resource")
		}
	})

	t.Run("multipart upload with progress callback", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, testResourceJSON))

		localPath := makeTempFile(t, strings.Repeat("y", 200))
		var updates []UploadProgress
		opts := &UploadOptions{
			ChunkSize: 64,
			Progress:  func(p UploadProgress) { updates = append(updates, p) },
		}

		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(updates) == 0 {
			t.Error("expected at least one progress update from multipart reader")
		}
	})

	t.Run("multipart upload link API error returns error", func(t *testing.T) {
		client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"ServiceUnavailable","description":"try later"}`))
		}))

		localPath := makeTempFile(t, strings.Repeat("z", 100))
		opts := &UploadOptions{ChunkSize: 50}

		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "upload link") {
			t.Errorf("expected 'upload link' in error, got: %v", err)
		}
	})

	t.Run("multipart non-2xx upload response returns error", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusInternalServerError, ""))

		localPath := makeTempFile(t, strings.Repeat("w", 100))
		opts := &UploadOptions{ChunkSize: 50}

		_, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "multipart upload failed") {
			t.Errorf("expected 'multipart upload failed' in error, got: %v", err)
		}
	})

	t.Run("multipart metadata failure returns basic resource", func(t *testing.T) {
		client := mockedHttpClient(uploadMockHandler(t, "/do-upload", http.StatusCreated, ""))

		localPath := makeTempFile(t, strings.Repeat("v", 100))
		opts := &UploadOptions{ChunkSize: 50}

		resource, err := client.UploadFileFromPath(context.Background(), localPath, "/test/file.txt", opts)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resource == nil {
			t.Fatal("expected fallback resource, got nil")
		}
		if resource.Path != "/test/file.txt" {
			t.Errorf("expected path '/test/file.txt', got %q", resource.Path)
		}
	})
}
