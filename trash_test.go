package disk

import (
	"context"
	"reflect"
	"testing"
)

func TestDeleteFromTrash(t *testing.T) {

	useCassette("/trash/delete")

	resp, errorResponse := client.DeleteFromTrash(context.Background(), TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	// when 204 OK
	if resp != nil {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}

func TestRestoreFromTrash(t *testing.T) {

	useCassette("trash/restore")

	resp, _, errorResponse := client.RestoreFromTrash(context.Background(), TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(link).Kind() != reflect.TypeOf(resp).Kind() {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}

func TestListTrashResources(t *testing.T) {

	useCassette("trash/list")

	resp, errorResponse := client.ListTrashResources(context.Background(), TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	trashResource := new(TrashResource)

	if reflect.TypeOf(trashResource).Kind() != reflect.TypeOf(resp).Kind() {
		t.Fatalf("error: expect %v, got %v", trashResource, resp)
	}
}
