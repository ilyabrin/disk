package disk_test

import (
	"context"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestTrashDelete(t *testing.T) {

	vcr := useCassette("trash/delete")
	defer vcr.Stop()

	resp, errorResponse := client.Trash.Delete(context.Background(), TEST_TRASH_FILE_PATH, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	// when 204 OK
	if resp != nil {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}

func TestTrashRestore(t *testing.T) {

	vcr := useCassette("trash/restore")
	defer vcr.Stop()

	resp, _, errorResponse := client.Trash.Restore(context.Background(), TEST_TRASH_FILE_PATH, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestTrashList(t *testing.T) {

	vcr := useCassette("trash/list")
	defer vcr.Stop()

	resp, errorResponse := client.Trash.List(context.Background(), "/", nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.TrashResource{}, t)
}
