package disk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// multipartUploadTimeout bounds a chunked upload when the caller did not set a
// deadline of their own.
const multipartUploadTimeout = 30 * time.Minute

// UploadProgress represents the progress of an upload operation
type UploadProgress struct {
	BytesUploaded int64
	TotalBytes    int64
	Percentage    float64
}

// ProgressCallback is called during upload to report progress
type ProgressCallback func(progress UploadProgress)

// UploadOptions contains options for file upload operations
type UploadOptions struct {
	Overwrite        bool             // Whether to overwrite existing files
	Progress         ProgressCallback // Optional progress callback
	ChunkSize        int64            // Size of chunks for multipart upload (0 = no chunking)
	ValidateChecksum bool             // Whether to validate file checksum after upload
}

// UploadFileFromPath uploads a file from the local filesystem to Yandex Disk
func (c *Client) UploadFileFromPath(ctx context.Context, localPath string, remotePath string, options *UploadOptions) (*Resource, error) {
	if localPath == "" {
		return nil, fmt.Errorf("local path cannot be empty")
	}
	if remotePath == "" {
		return nil, fmt.Errorf("remote path cannot be empty")
	}

	// Set default options if not provided
	if options == nil {
		options = &UploadOptions{}
	}

	c.Logger.Debug("Starting file upload from %s to %s", localPath, remotePath)

	// Step 1: Validate the local file
	fileInfo, err := c.validateLocalFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	c.Logger.Info("Uploading file: %s (size: %d bytes)", filepath.Base(localPath), fileInfo.Size())

	// Step 2: Check if we need multipart upload for large files
	if options.ChunkSize > 0 && fileInfo.Size() > options.ChunkSize {
		return c.uploadFileMultipart(ctx, localPath, remotePath, fileInfo, options)
	}

	// Step 3: Single file upload
	return c.uploadFileSingle(ctx, localPath, remotePath, fileInfo, options)
}

// validateLocalFile validates that the local file exists and is readable
func (c *Client) validateLocalFile(localPath string) (os.FileInfo, error) {
	// Check if file exists
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", localPath)
		}
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a file (not a directory)
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", localPath)
	}

	// Check if file is readable
	file, err := os.Open(localPath) // #nosec G304 -- localPath is supplied by the caller of this library
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("cannot close file: %w", err)
	}

	// Check file size (Yandex Disk has limits)
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("file is empty: %s", localPath)
	}

	c.Logger.Debug("File validation successful: %s (size: %d bytes)", localPath, fileInfo.Size())
	return fileInfo, nil
}

// uploadFileSingle handles single file upload without chunking
func (c *Client) uploadFileSingle(ctx context.Context, localPath string, remotePath string, fileInfo os.FileInfo, options *UploadOptions) (*Resource, error) {
	// Step 1: Get upload link from Yandex Disk API
	uploadLink, linkErr := c.GetLinkForUpload(ctx, remotePath)
	if linkErr != nil {
		c.Logger.LogError("get upload link", fmt.Errorf("failed to get upload link: %v", linkErr))
		return nil, fmt.Errorf("failed to get upload link: %v", linkErr)
	}

	if uploadLink == nil || uploadLink.Href == "" {
		return nil, fmt.Errorf("received invalid upload link")
	}

	c.Logger.Debug("Received upload link: %s", uploadLink.Href)

	// Step 2: Open the local file
	file, err := os.Open(localPath) // #nosec G304 -- localPath is supplied by the caller of this library
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Step 3: Create a progress reader if callback is provided
	var reader io.Reader = file
	if options.Progress != nil {
		reader = &progressReader{
			reader:   file,
			total:    fileInfo.Size(),
			callback: options.Progress,
		}
	}

	// Step 4: Create the HTTP request for file upload
	req, err := http.NewRequestWithContext(ctx, uploadLink.Method, uploadLink.Href, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set content type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = fileInfo.Size()

	// Handle overwrite policy
	if options.Overwrite {
		req.Header.Set("X-Overwrite", "true")
	}

	c.Logger.Debug("Uploading file with content type: %s", contentType)

	// Step 5: Execute the upload using the configured HTTP client
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.LogError("file upload", err)
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	// Step 6: Handle the response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	c.Logger.Info("File uploaded successfully: %s", remotePath)

	// Step 7: Get the uploaded resource metadata
	resource, metadataErr := c.GetMetadata(ctx, remotePath)
	if metadataErr != nil {
		c.Logger.Warn("Upload succeeded but failed to get resource metadata: %v", metadataErr)
		// Return a basic resource with the information we have
		return &Resource{
			Path: remotePath,
			Name: filepath.Base(localPath),
			Type: "file",
			Size: int(fileInfo.Size()),
		}, nil
	}

	return resource, nil
}

// uploadFileMultipart handles multipart upload for large files
func (c *Client) uploadFileMultipart(ctx context.Context, localPath string, remotePath string, fileInfo os.FileInfo, options *UploadOptions) (*Resource, error) {
	c.Logger.Info("Starting multipart upload for large file: %s (size: %d bytes)", localPath, fileInfo.Size())

	chunkSize := options.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 10 * 1024 * 1024 // Default 10MB chunks
	}

	totalChunks := (fileInfo.Size() + chunkSize - 1) / chunkSize
	c.Logger.Debug("Upload will be split into %d chunks of %d bytes each", totalChunks, chunkSize)

	// Note: Yandex Disk API doesn't have built-in resumable upload like Google Drive
	// For large files, we use the standard upload with better progress tracking and retry logic

	// Open the file for reading
	file, err := os.Open(localPath) // #nosec G304 -- localPath is supplied by the caller of this library
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get upload link
	uploadLink, linkErr := c.GetLinkForUpload(ctx, remotePath)
	if linkErr != nil {
		c.Logger.LogError("get upload link for multipart", fmt.Errorf("failed to get upload link: %v", linkErr))
		return nil, fmt.Errorf("failed to get upload link: %v", linkErr)
	}

	if uploadLink == nil || uploadLink.Href == "" {
		return nil, fmt.Errorf("received invalid upload link")
	}

	// Create a buffered reader for chunked progress tracking
	reader := &multipartProgressReader{
		reader:    file,
		total:     fileInfo.Size(),
		chunkSize: chunkSize,
		callback:  options.Progress,
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, uploadLink.Method, uploadLink.Href, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set headers
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = fileInfo.Size()

	if options.Overwrite {
		req.Header.Set("X-Overwrite", "true")
	}

	c.Logger.Debug("Starting multipart upload with content type: %s", contentType)

	// Large uploads need far more headroom than the default per-request timeout.
	// Use a client that shares this client's transport but is bounded by the
	// request context instead, so concurrent callers are not affected — mutating
	// c.HTTPClient.Timeout here would be a data race.
	uploadClient := &http.Client{
		Transport:     c.HTTPClient.Transport,
		CheckRedirect: c.HTTPClient.CheckRedirect,
		Jar:           c.HTTPClient.Jar,
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		uploadCtx, cancel := context.WithTimeout(ctx, multipartUploadTimeout)
		defer cancel()
		req = req.WithContext(uploadCtx)
	}

	resp, err := uploadClient.Do(req)
	if err != nil {
		c.Logger.LogError("multipart file upload", err)
		return nil, fmt.Errorf("multipart upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("multipart upload failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	c.Logger.Info("Multipart file uploaded successfully: %s", remotePath)

	// Get the uploaded resource metadata
	resource, metadataErr := c.GetMetadata(ctx, remotePath)
	if metadataErr != nil {
		c.Logger.Warn("Upload succeeded but failed to get resource metadata: %v", metadataErr)
		return &Resource{
			Path: remotePath,
			Name: filepath.Base(localPath),
			Type: "file",
			Size: int(fileInfo.Size()),
		}, nil
	}

	return resource, nil
}

// progressReader wraps an io.Reader to provide upload progress tracking
type progressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	if pr.callback != nil {
		percentage := float64(-1)
		if pr.total > 0 {
			percentage = float64(pr.current) / float64(pr.total) * 100
		}
		pr.callback(UploadProgress{
			BytesUploaded: pr.current,
			TotalBytes:    pr.total,
			Percentage:    percentage,
		})
	}

	return n, err
}

// multipartProgressReader provides chunked progress tracking for large file uploads
type multipartProgressReader struct {
	reader       io.Reader
	total        int64
	current      int64
	chunkSize    int64
	callback     ProgressCallback
	lastReported int64
}

func (mpr *multipartProgressReader) Read(p []byte) (int, error) {
	n, err := mpr.reader.Read(p)
	mpr.current += int64(n)

	if mpr.callback != nil {
		// Report progress every chunk or at the end
		if mpr.current-mpr.lastReported >= mpr.chunkSize || err == io.EOF {
			percentage := float64(-1)
			if mpr.total > 0 {
				percentage = float64(mpr.current) / float64(mpr.total) * 100
			}
			mpr.callback(UploadProgress{
				BytesUploaded: mpr.current,
				TotalBytes:    mpr.total,
				Percentage:    percentage,
			})
			mpr.lastReported = mpr.current
		}
	}

	return n, err
}

// DetectMimeType attempts to detect the MIME type of a file
func (c *Client) DetectMimeType(filePath string) (string, error) {
	// First try by extension
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType != "" {
		return mimeType, nil
	}

	// Try to detect from file content
	file, err := os.Open(filePath) // #nosec G304 -- filePath is supplied by the caller of this library
	if err != nil {
		return "", fmt.Errorf("cannot open file for MIME detection: %w", err)
	}
	defer file.Close()

	// Read up to the first 512 bytes for detection. io.ReadFull is used instead
	// of a bare Read so short reads on the first call do not truncate the sniff.
	buffer := make([]byte, 512)
	n, err := io.ReadFull(file, buffer)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("cannot read file for MIME detection: %w", err)
	}

	// Use http.DetectContentType
	detectedType := http.DetectContentType(buffer[:n])
	return detectedType, nil
}

// ValidateFilePath checks if a file path is valid for upload
func ValidateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains invalid character: %s", char)
		}
	}

	// Check path length (Yandex Disk limitation)
	if len(path) > 32768 {
		return fmt.Errorf("path too long (max 32768 characters)")
	}

	// Check if path starts with /
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}

	return nil
}

// UploadFileFromPathWithProgress is a convenience method for uploads with progress tracking
func (c *Client) UploadFileFromPathWithProgress(ctx context.Context, localPath string, remotePath string, overwrite bool, callback ProgressCallback) (*Resource, error) {
	options := &UploadOptions{
		Overwrite: overwrite,
		Progress:  callback,
	}
	return c.UploadFileFromPath(ctx, localPath, remotePath, options)
}

// UploadLargeFileFromPath is a convenience method for large file uploads with chunking
func (c *Client) UploadLargeFileFromPath(ctx context.Context, localPath string, remotePath string, chunkSizeMB int, callback ProgressCallback) (*Resource, error) {
	chunkSize := int64(chunkSizeMB) * 1024 * 1024 // Convert MB to bytes
	options := &UploadOptions{
		ChunkSize: chunkSize,
		Progress:  callback,
	}
	return c.UploadFileFromPath(ctx, localPath, remotePath, options)
}

// GetFileSize returns the size of a local file
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot get file size: %w", err)
	}
	return fileInfo.Size(), nil
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
