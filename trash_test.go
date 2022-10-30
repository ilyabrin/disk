package disk

import (
	"context"
	"reflect"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const TEST_TRASH_FILE_PATH = "trash:/___golang_API_dir_2_ddf8722d0aec88bfeb94a45a155511dbe151b764"

func TestDeleteFromTrash(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/trash/delete")
	if err != nil {
		t.Fatal(err)
	}
	defer vcr.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	vcr.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if vcr.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = vcr.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.DeleteFromTrash(ctx, TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	// when 204 OK
	if resp != nil {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}

func TestRestoreFromTrash(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/trash/restore")
	if err != nil {
		t.Fatal(err)
	}
	defer vcr.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	vcr.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if vcr.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = vcr.GetDefaultClient()

	ctx := context.Background()
	resp, _, errorResponse := client.RestoreFromTrash(ctx, TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(link).Kind() != reflect.TypeOf(resp).Kind() {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}

func TestListTrashResources(t *testing.T) {

	vcr, err := recorder.New(TEST_DATA_DIR + "/trash/list")
	if err != nil {
		t.Fatal(err)
	}
	defer vcr.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	vcr.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if vcr.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = vcr.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.ListTrashResources(ctx, TEST_TRASH_FILE_PATH, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	trashResource := new(TrashResource)

	if reflect.TypeOf(trashResource).Kind() != reflect.TypeOf(resp).Kind() {
		t.Fatalf("error: expect %v, got %v", trashResource, resp)
	}
}
