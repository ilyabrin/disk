package disk

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const TEST_DATA_DIR = "testdata/responses/"

// TODO: run before any test
const TEST_ACCESS_TOKEN = "test"

const TEST_DIR_NAME = "test_dir"

// TODO: fix it
func vcrTestClient(cassette_path string) (*recorder.Recorder, error) {
	rec, err := recorder.New(TEST_DATA_DIR + cassette_path)
	if err != nil {
		return nil, err
	}
	defer rec.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if rec.Mode() != recorder.ModeRecordOnce {
		return nil, errors.New("Recorder should be in ModeRecordOnce")
	}

	return rec, nil
}

func TestCreateDir(t *testing.T) {

	rec, err := recorder.New(TEST_DATA_DIR + "/disk/create_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Stop()

	hookDeleteToken := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(hookDeleteToken, recorder.AfterCaptureHook)

	if rec.Mode() != recorder.ModeRecordOnce {
		t.Fatal("Recorder should be in ModeRecordOnce")
	}

	client := New(TEST_ACCESS_TOKEN)
	client.HTTPClient = rec.GetDefaultClient()

	ctx := context.Background()
	resp, errorResponse := client.CreateDir(ctx, TEST_DIR_NAME, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if link == resp {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestDiskInfo(t *testing.T) {

	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/info")
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
	resp, errorResponse := client.DiskInfo(ctx, nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	disk := new(Disk)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(disk).Kind() {
		t.Fatalf("error: expect %v, got %v", disk, resp)
	}
}

func TestUpdateMetadata(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/update_meta")
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

	metadata := map[string]map[string]string{
		"custom_properties": {
			"key": "value",
			"foo": "bar",
		},
	}

	ctx := context.Background()
	resp, errorResponse := client.UpdateMetadata(ctx, "test_dir", metadata)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	resource := new(Resource)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(resource).Kind() {
		t.Fatalf("error: expect %v, got %v", resource, resp)
	}
}

func TestGetMetadata(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/get_meta")
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
	resp, errorResponse := client.GetMetadata(ctx, "test_dir", nil)

	t.Log(resp.CustomProperties["key"])

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	resource := new(Resource)
	if resource == resp {
		t.Fatalf("error: expect %v, got %v", resource, resp)
	}

	value := resp.CustomProperties["foo"]
	if value != "bar" {
		t.Fatalf("error: expect %v, got %v", value, "bar")
	}
}

func TestCopyResource(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/copy")
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
	resp, errorResponse := client.CopyResource(ctx, "test_dir", "test_dir_copy", nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetDownloadURL(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/download_url")
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
	resp, errorResponse := client.GetDownloadURL(ctx, "test_dir", nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

// TODO: fix timeout error
func TestGetSortedFiles(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/get_sorted_files")
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
	resp, errorResponse := client.GetSortedFiles(ctx, &optional_params{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	files := new(FilesResourceList)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestGetLastUploadedResources(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/last_uploaded")
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
	resp, errorResponse := client.GetLastUploadedResources(ctx, &optional_params{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	files := new(LastUploadedResourceList)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestMoveResource(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/move")
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
	resp, errorResponse := client.MoveResource(ctx, "test_dir_copy", "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetPublicResources(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/get_public_res")
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
	resp, errorResponse := client.GetPublicResources(ctx, &optional_params{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(PublicResourcesList)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestPublishResource(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/publish")
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
	resp, errorResponse := client.PublishResource(ctx, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUnpublishResource(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/unpublish")
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
	resp, errorResponse := client.UnpublishResource(ctx, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetLinkForUpload(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/get_upload_link")
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
	resp, errorResponse := client.GetLinkForUpload(ctx, "upload_path")

	if errorResponse != nil {
		t.Fatal(errorResponse)
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUploadFile(t *testing.T) {
	upload_link := "https://uploader7v.disk.yandex.net:443/upload-target/20221029T200308.792.utd.e8t7amr9zkrpoofffacoiggoz-k7v.6331006"

	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/upload_file")
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
	errorResponse := client.UploadFile(ctx, "LICENSE", upload_link, nil)

	if errorResponse.StatusCode != 201 {
		t.Fatalf("error: expect %v, got %v", 201, errorResponse.StatusCode)
	}

}

func TestDeleteResource(t *testing.T) {
	vcr, err := recorder.New(TEST_DATA_DIR + "/disk/delete_resource")
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
	resp := client.DeleteResource(ctx, "test_dir", false, nil)

	if nil != resp {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}
