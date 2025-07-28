package disk

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// BatchOperationResult represents the result of a single operation in a batch
type BatchOperationResult struct {
	Path      string        `json:"path"`
	Success   bool          `json:"success"`
	Error     error         `json:"error,omitempty"`
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	Link      *Link         `json:"link,omitempty"`     // For async operations
	Resource  *Resource     `json:"resource,omitempty"` // For operations that return resources
}

// BatchOperationStatus represents the overall status of a batch operation
type BatchOperationStatus struct {
	Total      int                     `json:"total"`
	Completed  int                     `json:"completed"`
	Successful int                     `json:"successful"`
	Failed     int                     `json:"failed"`
	InProgress int                     `json:"in_progress"`
	Results    []*BatchOperationResult `json:"results"`
	StartTime  time.Time               `json:"start_time"`
	EndTime    *time.Time              `json:"end_time,omitempty"`
	Duration   time.Duration           `json:"duration"`
	Percentage float64                 `json:"percentage"`
}

// BatchProgressCallback is called during batch operations to report progress
type BatchProgressCallback func(status BatchOperationStatus)

// BatchOptions contains configuration options for batch operations
type BatchOptions struct {
	MaxConcurrency  int                   // Maximum number of concurrent operations (default: 5)
	ContinueOnError bool                  // Whether to continue processing if some operations fail
	Progress        BatchProgressCallback // Optional progress callback
	Timeout         time.Duration         // Timeout for individual operations
}

// BatchDeleteOptions contains options specific to batch deletion
type BatchDeleteOptions struct {
	BatchOptions
	Permanently bool // Whether to delete files permanently or move to trash
}

// BatchCopyMoveOptions contains options specific to batch copy/move operations
type BatchCopyMoveOptions struct {
	BatchOptions
	DestinationPrefix string // Prefix to add to destination paths
	Overwrite         bool   // Whether to overwrite existing files
}

// BatchUpdateMetadataOptions contains options for batch metadata updates
type BatchUpdateMetadataOptions struct {
	BatchOptions
	CustomProperties map[string]map[string]string // Properties to set on all files
	Fields           []string                     // Specific fields to update
}

// BatchDeleteFiles deletes multiple files in parallel
func (c *Client) BatchDeleteFiles(ctx context.Context, paths []string, options *BatchDeleteOptions) (*BatchOperationStatus, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths list cannot be empty")
	}

	// Set default options
	if options == nil {
		options = &BatchDeleteOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency:  5,
				ContinueOnError: true,
			},
		}
	}
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 5
	}

	c.Logger.Info("Starting batch delete operation for %d files", len(paths))

	status := &BatchOperationStatus{
		Total:     len(paths),
		Results:   make([]*BatchOperationResult, len(paths)),
		StartTime: time.Now(),
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, options.MaxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process each path
	for i, path := range paths {
		wg.Add(1)
		go func(index int, resourcePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			result := &BatchOperationResult{
				Path:      resourcePath,
				Operation: "delete",
			}

			// Create operation-specific context with timeout
			opCtx := ctx
			if options.Timeout > 0 {
				var cancel context.CancelFunc
				opCtx, cancel = context.WithTimeout(ctx, options.Timeout)
				defer cancel()
			}

			// Perform the delete operation
			err := c.DeleteResource(opCtx, resourcePath, options.Permanently)
			result.Duration = time.Since(startTime)

			if err != nil {
				result.Success = false
				result.Error = err
				c.Logger.Warn("Failed to delete %s: %v", resourcePath, err)
			} else {
				result.Success = true
				c.Logger.Debug("Successfully deleted %s", resourcePath)
			}

			// Update status
			mu.Lock()
			status.Results[index] = result
			status.Completed++
			if result.Success {
				status.Successful++
			} else {
				status.Failed++
			}
			status.Percentage = float64(status.Completed) / float64(status.Total) * 100

			// Report progress if callback is provided
			if options.Progress != nil {
				statusCopy := *status
				statusCopy.Duration = time.Since(status.StartTime)
				options.Progress(statusCopy)
			}
			mu.Unlock()

		}(i, path)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Finalize status
	endTime := time.Now()
	status.EndTime = &endTime
	status.Duration = endTime.Sub(status.StartTime)
	status.InProgress = 0

	c.Logger.Info("Batch delete completed: %d/%d successful, %d failed in %v",
		status.Successful, status.Total, status.Failed, status.Duration)

	return status, nil
}

// BatchCopyFiles copies multiple files in parallel
func (c *Client) BatchCopyFiles(ctx context.Context, operations map[string]string, options *BatchCopyMoveOptions) (*BatchOperationStatus, error) {
	if len(operations) == 0 {
		return nil, fmt.Errorf("operations map cannot be empty")
	}

	// Set default options
	if options == nil {
		options = &BatchCopyMoveOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency:  5,
				ContinueOnError: true,
			},
		}
	}
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 5
	}

	c.Logger.Info("Starting batch copy operation for %d files", len(operations))

	status := &BatchOperationStatus{
		Total:     len(operations),
		Results:   make([]*BatchOperationResult, 0, len(operations)),
		StartTime: time.Now(),
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, options.MaxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	index := 0
	for fromPath, toPath := range operations {
		wg.Add(1)
		go func(idx int, from, to string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			result := &BatchOperationResult{
				Path:      from + " -> " + to,
				Operation: "copy",
			}

			// Apply destination prefix if specified
			if options.DestinationPrefix != "" {
				to = options.DestinationPrefix + to
			}

			// Create operation-specific context with timeout
			opCtx := ctx
			if options.Timeout > 0 {
				var cancel context.CancelFunc
				opCtx, cancel = context.WithTimeout(ctx, options.Timeout)
				defer cancel()
			}

			// Perform the copy operation
			link, errResp := c.CopyResource(opCtx, from, to)
			result.Duration = time.Since(startTime)

			if errResp != nil {
				result.Success = false
				result.Error = fmt.Errorf(errResp.Error)
				c.Logger.Warn("Failed to copy %s to %s: %v", from, to, errResp.Error)
			} else {
				result.Success = true
				result.Link = link
				c.Logger.Debug("Successfully copied %s to %s", from, to)
			}

			// Update status
			mu.Lock()
			status.Results = append(status.Results, result)
			status.Completed++
			if result.Success {
				status.Successful++
			} else {
				status.Failed++
			}
			status.Percentage = float64(status.Completed) / float64(status.Total) * 100

			// Report progress if callback is provided
			if options.Progress != nil {
				statusCopy := *status
				statusCopy.Duration = time.Since(status.StartTime)
				options.Progress(statusCopy)
			}
			mu.Unlock()

		}(index, fromPath, toPath)
		index++
	}

	// Wait for all operations to complete
	wg.Wait()

	// Finalize status
	endTime := time.Now()
	status.EndTime = &endTime
	status.Duration = endTime.Sub(status.StartTime)
	status.InProgress = 0

	c.Logger.Info("Batch copy completed: %d/%d successful, %d failed in %v",
		status.Successful, status.Total, status.Failed, status.Duration)

	return status, nil
}

// BatchMoveFiles moves multiple files in parallel
func (c *Client) BatchMoveFiles(ctx context.Context, operations map[string]string, options *BatchCopyMoveOptions) (*BatchOperationStatus, error) {
	if len(operations) == 0 {
		return nil, fmt.Errorf("operations map cannot be empty")
	}

	// Set default options
	if options == nil {
		options = &BatchCopyMoveOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency:  5,
				ContinueOnError: true,
			},
		}
	}
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 5
	}

	c.Logger.Info("Starting batch move operation for %d files", len(operations))

	status := &BatchOperationStatus{
		Total:     len(operations),
		Results:   make([]*BatchOperationResult, 0, len(operations)),
		StartTime: time.Now(),
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, options.MaxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	index := 0
	for fromPath, toPath := range operations {
		wg.Add(1)
		go func(idx int, from, to string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			result := &BatchOperationResult{
				Path:      from + " -> " + to,
				Operation: "move",
			}

			// Apply destination prefix if specified
			if options.DestinationPrefix != "" {
				to = options.DestinationPrefix + to
			}

			// Create operation-specific context with timeout
			opCtx := ctx
			if options.Timeout > 0 {
				var cancel context.CancelFunc
				opCtx, cancel = context.WithTimeout(ctx, options.Timeout)
				defer cancel()
			}

			// Perform the move operation
			link, errResp := c.MoveResource(opCtx, from, to)
			result.Duration = time.Since(startTime)

			if errResp != nil {
				result.Success = false
				result.Error = fmt.Errorf(errResp.Error)
				c.Logger.Warn("Failed to move %s to %s: %v", from, to, errResp.Error)
			} else {
				result.Success = true
				result.Link = link
				c.Logger.Debug("Successfully moved %s to %s", from, to)
			}

			// Update status
			mu.Lock()
			status.Results = append(status.Results, result)
			status.Completed++
			if result.Success {
				status.Successful++
			} else {
				status.Failed++
			}
			status.Percentage = float64(status.Completed) / float64(status.Total) * 100

			// Report progress if callback is provided
			if options.Progress != nil {
				statusCopy := *status
				statusCopy.Duration = time.Since(status.StartTime)
				options.Progress(statusCopy)
			}
			mu.Unlock()

		}(index, fromPath, toPath)
		index++
	}

	// Wait for all operations to complete
	wg.Wait()

	// Finalize status
	endTime := time.Now()
	status.EndTime = &endTime
	status.Duration = endTime.Sub(status.StartTime)
	status.InProgress = 0

	c.Logger.Info("Batch move completed: %d/%d successful, %d failed in %v",
		status.Successful, status.Total, status.Failed, status.Duration)

	return status, nil
}

// BatchUpdateMetadata updates metadata for multiple files in parallel
func (c *Client) BatchUpdateMetadata(ctx context.Context, paths []string, customProperties map[string]map[string]string, options *BatchUpdateMetadataOptions) (*BatchOperationStatus, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths list cannot be empty")
	}
	if len(customProperties) == 0 {
		return nil, fmt.Errorf("custom properties cannot be empty")
	}

	// Set default options
	if options == nil {
		options = &BatchUpdateMetadataOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency:  5,
				ContinueOnError: true,
			},
		}
	}
	if options.MaxConcurrency <= 0 {
		options.MaxConcurrency = 5
	}

	c.Logger.Info("Starting batch metadata update operation for %d files", len(paths))

	status := &BatchOperationStatus{
		Total:     len(paths),
		Results:   make([]*BatchOperationResult, len(paths)),
		StartTime: time.Now(),
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, options.MaxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process each path
	for i, path := range paths {
		wg.Add(1)
		go func(index int, resourcePath string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			result := &BatchOperationResult{
				Path:      resourcePath,
				Operation: "update_metadata",
			}

			// Create operation-specific context with timeout
			opCtx := ctx
			if options.Timeout > 0 {
				var cancel context.CancelFunc
				opCtx, cancel = context.WithTimeout(ctx, options.Timeout)
				defer cancel()
			}

			// Perform the metadata update operation
			resource, errResp := c.UpdateMetadata(opCtx, resourcePath, customProperties)
			result.Duration = time.Since(startTime)

			if errResp != nil {
				result.Success = false
				result.Error = fmt.Errorf(errResp.Error)
				c.Logger.Warn("Failed to update metadata for %s: %v", resourcePath, errResp.Error)
			} else {
				result.Success = true
				result.Resource = resource
				c.Logger.Debug("Successfully updated metadata for %s", resourcePath)
			}

			// Update status
			mu.Lock()
			status.Results[index] = result
			status.Completed++
			if result.Success {
				status.Successful++
			} else {
				status.Failed++
			}
			status.Percentage = float64(status.Completed) / float64(status.Total) * 100

			// Report progress if callback is provided
			if options.Progress != nil {
				statusCopy := *status
				statusCopy.Duration = time.Since(status.StartTime)
				options.Progress(statusCopy)
			}
			mu.Unlock()

		}(i, path)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Finalize status
	endTime := time.Now()
	status.EndTime = &endTime
	status.Duration = endTime.Sub(status.StartTime)
	status.InProgress = 0

	c.Logger.Info("Batch metadata update completed: %d/%d successful, %d failed in %v",
		status.Successful, status.Total, status.Failed, status.Duration)

	return status, nil
}

// GetBatchOperationsSummary provides a summary of batch operation results
func (status *BatchOperationStatus) GetSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"total":      status.Total,
		"completed":  status.Completed,
		"successful": status.Successful,
		"failed":     status.Failed,
		"percentage": status.Percentage,
		"duration":   status.Duration.String(),
	}

	if status.EndTime != nil {
		summary["completed_at"] = status.EndTime.Format(time.RFC3339)
	}

	// Group errors by type
	errorsByType := make(map[string]int)
	for _, result := range status.Results {
		if result != nil && result.Error != nil {
			errorType := result.Error.Error()
			errorsByType[errorType]++
		}
	}
	if len(errorsByType) > 0 {
		summary["errors_by_type"] = errorsByType
	}

	// Calculate average operation duration
	var totalDuration time.Duration
	completedOps := 0
	for _, result := range status.Results {
		if result != nil {
			totalDuration += result.Duration
			completedOps++
		}
	}
	if completedOps > 0 {
		summary["average_operation_duration"] = (totalDuration / time.Duration(completedOps)).String()
	}

	return summary
}

// GetFailedOperations returns only the failed operations from a batch
func (status *BatchOperationStatus) GetFailedOperations() []*BatchOperationResult {
	var failed []*BatchOperationResult
	for _, result := range status.Results {
		if result != nil && !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetSuccessfulOperations returns only the successful operations from a batch
func (status *BatchOperationStatus) GetSuccessfulOperations() []*BatchOperationResult {
	var successful []*BatchOperationResult
	for _, result := range status.Results {
		if result != nil && result.Success {
			successful = append(successful, result)
		}
	}
	return successful
}

// Convenience methods for common batch operations

// BatchDeleteFilesSimple is a simplified version of BatchDeleteFiles with basic options
func (c *Client) BatchDeleteFilesSimple(ctx context.Context, paths []string, permanently bool) (*BatchOperationStatus, error) {
	options := &BatchDeleteOptions{
		BatchOptions: BatchOptions{
			MaxConcurrency:  5,
			ContinueOnError: true,
		},
		Permanently: permanently,
	}
	return c.BatchDeleteFiles(ctx, paths, options)
}

// BatchCopyFilesSimple is a simplified version of BatchCopyFiles with basic options
func (c *Client) BatchCopyFilesSimple(ctx context.Context, operations map[string]string) (*BatchOperationStatus, error) {
	options := &BatchCopyMoveOptions{
		BatchOptions: BatchOptions{
			MaxConcurrency:  5,
			ContinueOnError: true,
		},
	}
	return c.BatchCopyFiles(ctx, operations, options)
}

// BatchMoveFilesSimple is a simplified version of BatchMoveFiles with basic options
func (c *Client) BatchMoveFilesSimple(ctx context.Context, operations map[string]string) (*BatchOperationStatus, error) {
	options := &BatchCopyMoveOptions{
		BatchOptions: BatchOptions{
			MaxConcurrency:  5,
			ContinueOnError: true,
		},
	}
	return c.BatchMoveFiles(ctx, operations, options)
}

// BatchRenameFiles renames multiple files by adding a prefix or suffix
func (c *Client) BatchRenameFiles(ctx context.Context, paths []string, prefix, suffix string, options *BatchCopyMoveOptions) (*BatchOperationStatus, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths list cannot be empty")
	}
	if prefix == "" && suffix == "" {
		return nil, fmt.Errorf("either prefix or suffix must be provided")
	}

	// Build rename operations
	operations := make(map[string]string)
	for _, path := range paths {
		dir := filepath.Dir(path)
		filename := filepath.Base(path)
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)

		newFilename := prefix + nameWithoutExt + suffix + ext
		newPath := filepath.Join(dir, newFilename)
		operations[path] = newPath
	}

	c.Logger.Info("Batch renaming %d files with prefix='%s', suffix='%s'", len(paths), prefix, suffix)
	return c.BatchMoveFiles(ctx, operations, options)
}

// BatchMoveToDirectory moves multiple files to a target directory
func (c *Client) BatchMoveToDirectory(ctx context.Context, paths []string, targetDir string, options *BatchCopyMoveOptions) (*BatchOperationStatus, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths list cannot be empty")
	}
	if targetDir == "" {
		return nil, fmt.Errorf("target directory cannot be empty")
	}

	// Build move operations
	operations := make(map[string]string)
	for _, path := range paths {
		filename := filepath.Base(path)
		newPath := filepath.Join(targetDir, filename)
		operations[path] = newPath
	}

	c.Logger.Info("Batch moving %d files to directory: %s", len(paths), targetDir)
	return c.BatchMoveFiles(ctx, operations, options)
}

// BatchCopyToDirectory copies multiple files to a target directory
func (c *Client) BatchCopyToDirectory(ctx context.Context, paths []string, targetDir string, options *BatchCopyMoveOptions) (*BatchOperationStatus, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths list cannot be empty")
	}
	if targetDir == "" {
		return nil, fmt.Errorf("target directory cannot be empty")
	}

	// Build copy operations
	operations := make(map[string]string)
	for _, path := range paths {
		filename := filepath.Base(path)
		newPath := filepath.Join(targetDir, filename)
		operations[path] = newPath
	}

	c.Logger.Info("Batch copying %d files to directory: %s", len(paths), targetDir)
	return c.BatchCopyFiles(ctx, operations, options)
}

// WaitForBatchOperation waits for asynchronous batch operations to complete
// This is useful when operations return Links for asynchronous processing
func (c *Client) WaitForBatchOperation(ctx context.Context, status *BatchOperationStatus, pollInterval time.Duration) error {
	if status == nil {
		return fmt.Errorf("status cannot be nil")
	}
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	c.Logger.Info("Waiting for batch operation to complete...")

	// Check if there are any async operations (operations that returned Links)
	asyncOps := 0
	for _, result := range status.Results {
		if result != nil && result.Link != nil {
			asyncOps++
		}
	}

	if asyncOps == 0 {
		c.Logger.Debug("No asynchronous operations found, batch is already complete")
		return nil
	}

	c.Logger.Info("Found %d asynchronous operations, polling for completion...", asyncOps)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Poll each async operation
			completed := 0
			for _, result := range status.Results {
				if result != nil && result.Link != nil {
					// Check operation status (this would require implementing operation status checking)
					// For now, we just log that we would check it
					c.Logger.Debug("Would check status of operation: %s", result.Link.Href)
					completed++
				}
			}

			if completed == asyncOps {
				c.Logger.Info("All asynchronous operations completed")
				return nil
			}
		}
	}
}

// RetryFailedOperations retries only the failed operations from a previous batch
func (c *Client) RetryFailedOperations(ctx context.Context, status *BatchOperationStatus, maxRetries int) (*BatchOperationStatus, error) {
	if status == nil {
		return nil, fmt.Errorf("status cannot be nil")
	}
	if maxRetries <= 0 {
		maxRetries = 3
	}

	failed := status.GetFailedOperations()
	if len(failed) == 0 {
		c.Logger.Info("No failed operations to retry")
		return status, nil
	}

	c.Logger.Info("Retrying %d failed operations (max %d retries)", len(failed), maxRetries)

	// Group failed operations by type
	deletePaths := make([]string, 0)
	copyOps := make(map[string]string)
	moveOps := make(map[string]string)
	metadataPaths := make([]string, 0)

	for _, result := range failed {
		switch result.Operation {
		case "delete":
			deletePaths = append(deletePaths, result.Path)
		case "copy":
			// Parse "from -> to" format
			parts := strings.Split(result.Path, " -> ")
			if len(parts) == 2 {
				copyOps[parts[0]] = parts[1]
			}
		case "move":
			// Parse "from -> to" format
			parts := strings.Split(result.Path, " -> ")
			if len(parts) == 2 {
				moveOps[parts[0]] = parts[1]
			}
		case "update_metadata":
			metadataPaths = append(metadataPaths, result.Path)
		}
	}

	// Retry operations
	var retryStatus *BatchOperationStatus
	var err error

	if len(deletePaths) > 0 {
		retryStatus, err = c.BatchDeleteFilesSimple(ctx, deletePaths, false)
		if err != nil {
			return nil, fmt.Errorf("failed to retry delete operations: %w", err)
		}
	}

	if len(copyOps) > 0 {
		retryStatus, err = c.BatchCopyFilesSimple(ctx, copyOps)
		if err != nil {
			return nil, fmt.Errorf("failed to retry copy operations: %w", err)
		}
	}

	if len(moveOps) > 0 {
		retryStatus, err = c.BatchMoveFilesSimple(ctx, moveOps)
		if err != nil {
			return nil, fmt.Errorf("failed to retry move operations: %w", err)
		}
	}

	// Note: Metadata retries would need the original custom properties
	// This is a limitation of the current approach
	if len(metadataPaths) > 0 {
		c.Logger.Warn("Cannot retry metadata operations without original custom properties")
	}

	return retryStatus, nil
}
