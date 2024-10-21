package disk_test

import (
	"context"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestOperationStatus(t *testing.T) {

	vcr := useCassette("operation/status")
	defer vcr.Stop()

	resp, errorResponse := client.Operation.Status(context.Background(), "8c6f3a7c126a0f966476c141514951d0472e45819157cff9e88185f132d1e6b8", nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	if !disk.InArray(resp.Status, []string{"success, in-progress", "failed"}) {
		t.Fatal("Operation status error")
	}
	checkTypes(resp, &disk.Operation{}, t)
}
