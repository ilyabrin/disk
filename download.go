package disk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadProgress represents the progress of a download operation.
type DownloadProgress struct {
	BytesDownloaded int64
	TotalBytes      int64   // -1 if content length is unknown
	Percentage      float64 // 0–100, or -1 if total is unknown
}

// DownloadOptions contains options for file download operations.
type DownloadOptions struct {
	Progress  DownloadProgressCallback // Optional progress callback
	Overwrite bool                     // Whether to overwrite an existing local file
}

// DownloadProgressCallback is called during download to report progress.
type DownloadProgressCallback func(progress DownloadProgress)

// DownloadFileToPath downloads a file from Yandex Disk to the local filesystem.
// remotePath is the path on Yandex Disk (e.g. "/Photos/image.jpg").
// localPath is the destination path on the local filesystem.
func (c *Client) DownloadFileToPath(ctx context.Context, remotePath string, localPath string, options *DownloadOptions) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}
	if localPath == "" {
		return fmt.Errorf("local path cannot be empty")
	}
	if options == nil {
		options = &DownloadOptions{}
	}

	// Check if local file already exists
	if !options.Overwrite {
		if _, err := os.Stat(localPath); err == nil {
			return fmt.Errorf("local file already exists: %s (use Overwrite option to replace)", localPath)
		}
	}

	c.Logger.Debug("Starting download: %s -> %s", remotePath, localPath)

	// Step 1: Get a temporary download URL from the API
	link, errResp := c.GetDownloadURL(ctx, remotePath)
	if errResp != nil {
		return fmt.Errorf("failed to get download URL: %s", errResp.Error)
	}
	if link == nil || link.Href == "" {
		return fmt.Errorf("received invalid download link")
	}

	c.Logger.Debug("Received download link: %s", link.Href)

	// Step 2: Execute the download request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.Href, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.LogError("file download", err)
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Step 3: Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0o750); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Step 4: Create the local file
	file, err := os.Create(localPath) // #nosec G304 -- localPath is supplied by the caller of this library
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Step 5: Copy with optional progress tracking
	totalBytes := resp.ContentLength // -1 if unknown
	var written int64

	buf := make([]byte, 32*1024)
	for {
		nr, readErr := resp.Body.Read(buf)
		if nr > 0 {
			nw, writeErr := file.Write(buf[:nr])
			written += int64(nw)
			if writeErr != nil {
				return fmt.Errorf("failed to write to local file: %w", writeErr)
			}

			if options.Progress != nil {
				var pct float64 = -1
				if totalBytes > 0 {
					pct = float64(written) / float64(totalBytes) * 100
				}
				options.Progress(DownloadProgress{
					BytesDownloaded: written,
					TotalBytes:      totalBytes,
					Percentage:      pct,
				})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to read download stream: %w", readErr)
		}
	}

	c.Logger.Info("File downloaded successfully: %s -> %s (%s)", remotePath, localPath, FormatFileSize(written))
	return nil
}

// DownloadFileToPathWithProgress is a convenience wrapper for DownloadFileToPath.
func (c *Client) DownloadFileToPathWithProgress(ctx context.Context, remotePath string, localPath string, overwrite bool, callback DownloadProgressCallback) error {
	options := &DownloadOptions{
		Overwrite: overwrite,
		Progress:  callback,
	}
	return c.DownloadFileToPath(ctx, remotePath, localPath, options)
}
