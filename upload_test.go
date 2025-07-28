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
			bytes   int64
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