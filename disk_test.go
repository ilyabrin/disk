package disk

import (
	"context"
	"reflect"
	"testing"
)

func TestCreateDir(t *testing.T) {

	useCassette("/disk/create_dir")

	ctx := context.Background()
	resp, errorResponse := client.CreateDir(ctx, TEST_DIR_NAME, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestDiskInfo(t *testing.T) {

	useCassette("/disk/info")

	resp, errorResponse := client.DiskInfo(context.Background(), nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	disk := new(Disk)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(disk).Kind() {
		t.Fatalf("error: expect %v, got %v", disk, resp)
	}

	if client.req_url != client.api_url {
		t.Fatalf("error: expect %v, got %v", client.req_url, client.api_url)
	}
}

func TestUpdateMetadata(t *testing.T) {

	useCassette("/disk/update_meta")

	metadata := map[string]map[string]string{
		"custom_properties": {
			"key": "value",
			"foo": "bar",
		},
	}

	ctx := context.Background()
	resp, errorResponse := client.UpdateMetadata(ctx, TEST_DIR_NAME, metadata)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	resource := new(Resource)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(resource).Kind() {
		t.Fatalf("error: expect %v, got %v", resource, resp)
	}
}

func TestGetMetadata(t *testing.T) {

	useCassette("/disk/get_meta")

	ctx := context.Background()
	resp, errorResponse := client.GetMetadata(ctx, TEST_DIR_NAME, nil)

	resource := new(Resource)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(resource).Kind() {
		t.Fatalf("error: expect %v, got %v", resource, resp)
	}

	value := resp.CustomProperties["foo"]
	if value != "bar" {
		t.Fatalf("error: expect %v, got %v", value, "bar")
	}
}

func TestCopyResource(t *testing.T) {

	useCassette("/disk/copy")

	resp, errorResponse := client.CopyResource(context.Background(), TEST_DIR_NAME, TEST_DIR_NAME_COPY, nil)

	// TODO: refactor
	expect := "https://cloud-api.yandex.net/v1/disk/resources/copy?from=test_dir&path=test_dir_copy"
	got := client.req_url
	if got != expect {
		t.Fatalf("error: expect %v, got %v", expect, got)
	}

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetDownloadURL(t *testing.T) {

	useCassette("/disk/download_url")

	resp, errorResponse := client.GetDownloadURL(context.Background(), TEST_DIR_NAME, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetSortedFiles(t *testing.T) {

	useCassette("/disk/get_sorted_files")

	resp, errorResponse := client.GetSortedFiles(context.Background(), &queryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	files := new(FilesResourceList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestGetLastUploadedResources(t *testing.T) {

	useCassette("/disk/last_uploaded")

	resp, errorResponse := client.GetLastUploadedResources(context.Background(), &queryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	files := new(LastUploadedResourceList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestMoveResource(t *testing.T) {

	useCassette("/disk/move")

	resp, errorResponse := client.MoveResource(context.Background(), TEST_DIR_NAME_COPY, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetPublicResources(t *testing.T) {

	useCassette("/disk/get_public_res")

	resp, errorResponse := client.GetPublicResources(context.Background(), &queryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(PublicResourcesList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestPublishResource(t *testing.T) {
	useCassette("/disk/publish")

	resp, errorResponse := client.PublishResource(context.Background(), "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUnpublishResource(t *testing.T) {

	useCassette("/disk/unpublish")
	ctx := context.Background()
	resp, errorResponse := client.UnpublishResource(ctx, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetLinkForUpload(t *testing.T) {

	useCassette("/disk/get_upload_link")

	resp, errorResponse := client.GetLinkForUpload(context.Background(), "upload_path")

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUploadFile(t *testing.T) {

	upload_link := "https://uploader7v.disk.yandex.net:443/upload-target/20221029T200308.792.utd.e8t7amr9zkrpoofffacoiggoz-k7v.6331006"

	useCassette("/disk/upload_file")

	errorResponse := client.UploadFile(context.Background(), "LICENSE", upload_link, nil)

	if errorResponse.StatusCode != 201 {
		t.Fatalf("error: expect %v, got %v", 201, errorResponse.StatusCode)
	}

}

func TestDeleteResource(t *testing.T) {
	useCassette("/disk/delete_resource")

	resp := client.DeleteResource(context.Background(), TEST_DIR_NAME, false, nil)

	if nil != resp {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}
