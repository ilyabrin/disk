package disk

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBatchDeleteFiles(t *testing.T) {
	t.Run("BatchDeleteFiles validates empty paths", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.BatchDeleteFiles(context.Background(), []string{}, nil)
		if err == nil {
			t.Error("Expected error for empty paths list")
		}
		if !strings.Contains(err.Error(), "paths list cannot be empty") {
			t.Errorf("Expected 'paths list cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("BatchDeleteFiles with default options", func(t *testing.T) {
		client, _ := New("test-token")

		paths := []string{"/file1.txt", "/file2.txt"}
		
		// This will fail at the API level but we're testing the batch structure
		status, err := client.BatchDeleteFiles(context.Background(), paths, nil)
		if err != nil {
			t.Fatal("Batch delete should not fail on setup:", err)
		}

		if status == nil {
			t.Fatal("Expected status to be returned")
		}

		if status.Total != 2 {
			t.Errorf("Expected total 2, got %d", status.Total)
		}

		if len(status.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(status.Results))
		}
	})

	t.Run("BatchDeleteFiles with custom options", func(t *testing.T) {
		client, _ := New("test-token")

		options := &BatchDeleteOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency:  2,
				ContinueOnError: false,
				Timeout:         5 * time.Second,
			},
			Permanently: true,
		}

		paths := []string{"/file1.txt"}
		
		status, err := client.BatchDeleteFiles(context.Background(), paths, options)
		if err != nil {
			t.Fatal("Batch delete should not fail on setup:", err)
		}

		if status.Total != 1 {
			t.Errorf("Expected total 1, got %d", status.Total)
		}
	})
}

func TestBatchCopyFiles(t *testing.T) {
	t.Run("BatchCopyFiles validates empty operations", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.BatchCopyFiles(context.Background(), map[string]string{}, nil)
		if err == nil {
			t.Error("Expected error for empty operations map")
		}
		if !strings.Contains(err.Error(), "operations map cannot be empty") {
			t.Errorf("Expected 'operations map cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("BatchCopyFiles with operations", func(t *testing.T) {
		client, _ := New("test-token")

		operations := map[string]string{
			"/source1.txt": "/dest1.txt",
			"/source2.txt": "/dest2.txt",
		}
		
		status, err := client.BatchCopyFiles(context.Background(), operations, nil)
		if err != nil {
			t.Fatal("Batch copy should not fail on setup:", err)
		}

		if status == nil {
			t.Fatal("Expected status to be returned")
		}

		if status.Total != 2 {
			t.Errorf("Expected total 2, got %d", status.Total)
		}

		if len(status.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(status.Results))
		}
	})

	t.Run("BatchCopyFiles with destination prefix", func(t *testing.T) {
		client, _ := New("test-token")

		options := &BatchCopyMoveOptions{
			BatchOptions: BatchOptions{
				MaxConcurrency: 3,
			},
			DestinationPrefix: "/backup",
		}

		operations := map[string]string{
			"/source.txt": "/dest.txt",
		}
		
		status, err := client.BatchCopyFiles(context.Background(), operations, options)
		if err != nil {
			t.Fatal("Batch copy should not fail on setup:", err)
		}

		if status.Total != 1 {
			t.Errorf("Expected total 1, got %d", status.Total)
		}
	})
}

func TestBatchMoveFiles(t *testing.T) {
	t.Run("BatchMoveFiles validates empty operations", func(t *testing.T) {
		client, _ := New("test-token")

		_, err := client.BatchMoveFiles(context.Background(), map[string]string{}, nil)
		if err == nil {
			t.Error("Expected error for empty operations map")
		}
		if !strings.Contains(err.Error(), "operations map cannot be empty") {
			t.Errorf("Expected 'operations map cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("BatchMoveFiles with operations", func(t *testing.T) {
		client, _ := New("test-token")

		operations := map[string]string{
			"/old1.txt": "/new1.txt",
			"/old2.txt": "/new2.txt",
		}
		
		status, err := client.BatchMoveFiles(context.Background(), operations, nil)
		if err != nil {
			t.Fatal("Batch move should not fail on setup:", err)
		}

		if status == nil {
			t.Fatal("Expected status to be returned")
		}

		if status.Total != 2 {
			t.Errorf("Expected total 2, got %d", status.Total)
		}
	})
}

func TestBatchUpdateMetadata(t *testing.T) {
	t.Run("BatchUpdateMetadata validates empty paths", func(t *testing.T) {
		client, _ := New("test-token")

		customProps := map[string]map[string]string{
			"custom": {
				"tag": "important",
			},
		}

		_, err := client.BatchUpdateMetadata(context.Background(), []string{}, customProps, nil)
		if err == nil {
			t.Error("Expected error for empty paths list")
		}
		if !strings.Contains(err.Error(), "paths list cannot be empty") {
			t.Errorf("Expected 'paths list cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("BatchUpdateMetadata validates empty properties", func(t *testing.T) {
		client, _ := New("test-token")

		paths := []string{"/file1.txt"}
		customProps := map[string]map[string]string{}

		_, err := client.BatchUpdateMetadata(context.Background(), paths, customProps, nil)
		if err == nil {
			t.Error("Expected error for empty custom properties")
		}
		if !strings.Contains(err.Error(), "custom properties cannot be empty") {
			t.Errorf("Expected 'custom properties cannot be empty' error, got: %s", err.Error())
		}
	})

	t.Run("BatchUpdateMetadata with valid inputs", func(t *testing.T) {
		client, _ := New("test-token")

		paths := []string{"/file1.txt", "/file2.txt"}
		customProps := map[string]map[string]string{
			"custom": {
				"tag":    "important",
				"author": "user123",
			},
		}
		
		status, err := client.BatchUpdateMetadata(context.Background(), paths, customProps, nil)
		if err != nil {
			t.Fatal("Batch metadata update should not fail on setup:", err)
		}

		if status == nil {
			t.Fatal("Expected status to be returned")
		}

		if status.Total != 2 {
			t.Errorf("Expected total 2, got %d", status.Total)
		}

		if len(status.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(status.Results))
		}
	})
}

func TestBatchOperationStatus(t *testing.T) {
	t.Run("GetSummary provides correct summary", func(t *testing.T) {
		endTime := time.Now()
		status := &BatchOperationStatus{
			Total:       5,
			Completed:   5,
			Successful:  3,
			Failed:      2,
			Percentage:  100.0,
			Duration:    time.Minute,
			EndTime:     &endTime,
			Results: []*BatchOperationResult{
				{Path: "/file1.txt", Success: true, Operation: "delete", Duration: time.Second},
				{Path: "/file2.txt", Success: true, Operation: "delete", Duration: time.Second},
				{Path: "/file3.txt", Success: true, Operation: "delete", Duration: time.Second},
				{Path: "/file4.txt", Success: false, Error: errors.New("error1"), Operation: "delete", Duration: time.Second},
				{Path: "/file5.txt", Success: false, Error: errors.New("error1"), Operation: "delete", Duration: time.Second},
			},
		}

		summary := status.GetSummary()

		if summary["total"].(int) != 5 {
			t.Errorf("Expected total 5, got %v", summary["total"])
		}
		if summary["successful"].(int) != 3 {
			t.Errorf("Expected successful 3, got %v", summary["successful"])
		}
		if summary["failed"].(int) != 2 {
			t.Errorf("Expected failed 2, got %v", summary["failed"])
		}
		if summary["percentage"].(float64) != 100.0 {
			t.Errorf("Expected percentage 100.0, got %v", summary["percentage"])
		}
	})

	t.Run("GetFailedOperations returns only failed operations", func(t *testing.T) {
		status := &BatchOperationStatus{
			Results: []*BatchOperationResult{
				{Path: "/file1.txt", Success: true, Operation: "delete"},
				{Path: "/file2.txt", Success: false, Error: errors.New("error"), Operation: "delete"},
				{Path: "/file3.txt", Success: false, Error: errors.New("error"), Operation: "delete"},
			},
		}

		failed := status.GetFailedOperations()
		if len(failed) != 2 {
			t.Errorf("Expected 2 failed operations, got %d", len(failed))
		}

		for _, result := range failed {
			if result.Success {
				t.Error("GetFailedOperations returned a successful operation")
			}
		}
	})

	t.Run("GetSuccessfulOperations returns only successful operations", func(t *testing.T) {
		status := &BatchOperationStatus{
			Results: []*BatchOperationResult{
				{Path: "/file1.txt", Success: true, Operation: "delete"},
				{Path: "/file2.txt", Success: false, Error: errors.New("error"), Operation: "delete"},
				{Path: "/file3.txt", Success: true, Operation: "delete"},
			},
		}

		successful := status.GetSuccessfulOperations()
		if len(successful) != 2 {
			t.Errorf("Expected 2 successful operations, got %d", len(successful))
		}

		for _, result := range successful {
			if !result.Success {
				t.Error("GetSuccessfulOperations returned a failed operation")
			}
		}
	})
}

func TestBatchOptions(t *testing.T) {
	t.Run("Default options are applied correctly", func(t *testing.T) {
		client, _ := New("test-token")

		paths := []string{"/file1.txt"}
		
		// Test with nil options - should use defaults
		status, err := client.BatchDeleteFiles(context.Background(), paths, nil)
		if err != nil {
			t.Fatal("Batch delete should not fail on setup:", err)
		}

		if status.Total != 1 {
			t.Errorf("Expected total 1, got %d", status.Total)
		}
	})

	t.Run("Progress callback structure", func(t *testing.T) {
		client, _ := New("test-token")

		var progressUpdates []BatchOperationStatus
		progressCallback := func(status BatchOperationStatus) {
			progressUpdates = append(progressUpdates, status)
		}

		options := &BatchDeleteOptions{
			BatchOptions: BatchOptions{
				Progress: progressCallback,
			},
		}

		paths := []string{"/file1.txt"}
		
		status, err := client.BatchDeleteFiles(context.Background(), paths, options)
		if err != nil {
			t.Fatal("Batch delete should not fail on setup:", err)
		}

		if status.Total != 1 {
			t.Errorf("Expected total 1, got %d", status.Total)
		}

		// Progress updates will happen during actual operations
		// Here we just verify the callback structure is correct
	})
}

func TestBatchConvenienceMethods(t *testing.T) {
	t.Run("BatchDeleteFilesSimple uses correct defaults", func(t *testing.T) {
		client, _ := New("test-token")

		paths := []string{"/file1.txt"}
		
		status, err := client.BatchDeleteFilesSimple(context.Background(), paths, true)
		if err != nil {
			t.Fatal("Batch delete simple should not fail on setup:", err)
		}

		if status.Total != 1 {
			t.Errorf("Expected total 1, got %d", status.Total)
		}
	})

	t.Run("BatchRenameFiles validates inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test empty paths
		_, err := client.BatchRenameFiles(context.Background(), []string{}, "prefix_", "", nil)
		if err == nil {
			t.Error("Expected error for empty paths list")
		}

		// Test empty prefix and suffix
		_, err = client.BatchRenameFiles(context.Background(), []string{"/file.txt"}, "", "", nil)
		if err == nil {
			t.Error("Expected error when both prefix and suffix are empty")
		}
		if !strings.Contains(err.Error(), "either prefix or suffix must be provided") {
			t.Errorf("Expected specific error message, got: %s", err.Error())
		}
	})

	t.Run("BatchMoveToDirectory validates inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test empty paths
		_, err := client.BatchMoveToDirectory(context.Background(), []string{}, "/target", nil)
		if err == nil {
			t.Error("Expected error for empty paths list")
		}

		// Test empty target directory
		_, err = client.BatchMoveToDirectory(context.Background(), []string{"/file.txt"}, "", nil)
		if err == nil {
			t.Error("Expected error for empty target directory")
		}
		if !strings.Contains(err.Error(), "target directory cannot be empty") {
			t.Errorf("Expected specific error message, got: %s", err.Error())
		}
	})

	t.Run("BatchCopyToDirectory validates inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test empty paths
		_, err := client.BatchCopyToDirectory(context.Background(), []string{}, "/target", nil)
		if err == nil {
			t.Error("Expected error for empty paths list")
		}

		// Test empty target directory
		_, err = client.BatchCopyToDirectory(context.Background(), []string{"/file.txt"}, "", nil)
		if err == nil {
			t.Error("Expected error for empty target directory")
		}
	})
}

func TestBatchUtilityMethods(t *testing.T) {
	t.Run("WaitForBatchOperation validates inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test nil status
		err := client.WaitForBatchOperation(context.Background(), nil, time.Second)
		if err == nil {
			t.Error("Expected error for nil status")
		}
		if !strings.Contains(err.Error(), "status cannot be nil") {
			t.Errorf("Expected specific error message, got: %s", err.Error())
		}
	})

	t.Run("WaitForBatchOperation handles no async operations", func(t *testing.T) {
		client, _ := New("test-token")

		status := &BatchOperationStatus{
			Results: []*BatchOperationResult{
				{Path: "/file1.txt", Success: true, Operation: "delete"},
			},
		}

		err := client.WaitForBatchOperation(context.Background(), status, time.Second)
		if err != nil {
			t.Errorf("Expected no error for synchronous operations, got: %s", err.Error())
		}
	})

	t.Run("RetryFailedOperations validates inputs", func(t *testing.T) {
		client, _ := New("test-token")

		// Test nil status
		_, err := client.RetryFailedOperations(context.Background(), nil, 3)
		if err == nil {
			t.Error("Expected error for nil status")
		}
		if !strings.Contains(err.Error(), "status cannot be nil") {
			t.Errorf("Expected specific error message, got: %s", err.Error())
		}
	})

	t.Run("RetryFailedOperations handles no failed operations", func(t *testing.T) {
		client, _ := New("test-token")

		status := &BatchOperationStatus{
			Results: []*BatchOperationResult{
				{Path: "/file1.txt", Success: true, Operation: "delete"},
			},
		}

		retryStatus, err := client.RetryFailedOperations(context.Background(), status, 3)
		if err != nil {
			t.Errorf("Expected no error when no failed operations, got: %s", err.Error())
		}
		if retryStatus != status {
			t.Error("Expected original status to be returned when no failed operations")
		}
	})
}