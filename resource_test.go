package disk_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestCreateDir(t *testing.T) {

	UseCassette("disk/create_dir")

	ctx := context.Background()
	resp, errorResponse := client.CreateDir(ctx, TEST_DIR_NAME, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUpdateMetadata(t *testing.T) {

	UseCassette("disk/update_meta")

	metadata := &disk.Metadata{
		"custom_properties": map[string]string{
			"key": "value",
			"foo": "bar",
		},
	}

	resp, errorResponse := client.UpdateMetadata(context.Background(), TEST_DIR_NAME, metadata)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	resource := new(disk.Resource)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(resource).Kind() {
		t.Fatalf("error: expect %v, got %v", resource, resp)
	}
}

func TestGetMetadata(t *testing.T) {

	UseCassette("disk/get_meta")

	resp, errorResponse := client.GetMetadata(context.Background(), TEST_DIR_NAME, nil)

	resource := new(disk.Resource)

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

	UseCassette("disk/copy")

	resp, errorResponse := client.CopyResource(context.Background(), TEST_DIR_NAME, TEST_DIR_NAME_COPY, nil)

	// TODO: refactor
	expect := "https://cloud-api.yandex.net/v1/disk/resources/copy?from=test_dir&path=test_dir_copy"
	got := client.ReqURL()
	if got != expect {
		t.Fatalf("error: expect %v, got %v", expect, got)
	}

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetDownloadURL(t *testing.T) {

	UseCassette("disk/download_url")

	resp, errorResponse := client.GetDownloadURL(context.Background(), TEST_DIR_NAME, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetSortedFiles(t *testing.T) {

	UseCassette("disk/get_sorted_files")

	resp, errorResponse := client.GetSortedFiles(context.Background(), &disk.QueryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	files := new(disk.FilesResourceList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestGetLastUploadedResources(t *testing.T) {

	UseCassette("disk/last_uploaded")

	resp, errorResponse := client.GetLastUploadedResources(context.Background(), &disk.QueryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	files := new(disk.LastUploadedResourceList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(files).Kind() {
		t.Fatalf("error: expect %v, got %v", files, resp)
	}
}

func TestMoveResource(t *testing.T) {

	UseCassette("disk/move")

	resp, errorResponse := client.MoveResource(context.Background(), TEST_DIR_NAME_COPY, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetPublicResources(t *testing.T) {

	UseCassette("disk/get_public_res")

	resp, errorResponse := client.GetPublicResources(context.Background(), &disk.QueryParams{
		"limit": "1",
	})

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.PublicResourcesList)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestPublishResource(t *testing.T) {

	UseCassette("disk/publish")

	resp, errorResponse := client.PublishResource(context.Background(), "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUnpublishResource(t *testing.T) {

	UseCassette("disk/unpublish")
	ctx := context.Background()
	resp, errorResponse := client.UnpublishResource(ctx, "test_dir_moved", nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)
	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestGetLinkForUpload(t *testing.T) {

	UseCassette("disk/get_upload_link")

	resp, errorResponse := client.GetLinkForUpload(context.Background(), "upload_path")

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestUploadFile(t *testing.T) {

	upload_link := "https://uploader7v.disk.yandex.net:443/upload-target/20221029T200308.792.utd.e8t7amr9zkrpoofffacoiggoz-k7v.6331006"

	UseCassette("disk/upload_file")

	errorResponse := client.UploadFile(context.Background(), "LICENSE", upload_link, nil)

	if errorResponse.StatusCode != http.StatusCreated {
		t.Fatalf("error: expect %v, got %v", 201, errorResponse.StatusCode)
	}

}

func TestDeleteResource(t *testing.T) {

	UseCassette("disk/delete_resource")

	resp := client.DeleteResource(context.Background(), TEST_DIR_NAME, false, nil)

	if nil != resp {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}
