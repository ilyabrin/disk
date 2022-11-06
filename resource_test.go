package disk_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestCreateDir(t *testing.T) {

	vcr := useCassette("resources/create_dir")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.CreateDir(context.Background(), TEST_DIR_NAME, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestUpdateMetadata(t *testing.T) {

	vcr := useCassette("resources/update_meta")
	defer vcr.Stop()

	metadata := &disk.Metadata{
		"custom_properties": map[string]string{
			"key": "value",
			"foo": "bar",
		},
	}

	resp, errorResponse := client.Resources.UpdateMeta(context.Background(), TEST_DIR_NAME, metadata)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Resource{}, t)
}

func TestGetMetadata(t *testing.T) {

	vcr := useCassette("resources/get_meta")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.Meta(context.Background(), TEST_DIR_NAME, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Resource{}, t)

	value := resp.CustomProperties["foo"]
	if value != "bar" {
		t.Fatalf("error: expect %v, got %v", value, "bar")
	}
}

func TestCopyResource(t *testing.T) {

	vcr := useCassette("resources/copy")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.Copy(context.Background(), TEST_DIR_NAME, TEST_DIR_NAME_COPY, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	expect := client.ApiURL() + "resources/copy?from=test_dir&path=test_dir_copy"
	if expect != client.ReqURL() {
		t.Fatalf("error: expect %v, got %v", expect, client.ReqURL())
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestGetDownloadURL(t *testing.T) {

	vcr := useCassette("resources/download_url")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.DownloadURL(context.Background(), TEST_DIR_NAME, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestGetSortedFiles(t *testing.T) {

	vcr := useCassette("resources/get_sorted_files")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.GetSortedFiles(context.Background(), &disk.QueryParams{"limit": "1"})
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.FilesResourceList{}, t)
}

func TestGetLastUploadedResources(t *testing.T) {

	vcr := useCassette("resources/last_uploaded")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.ListLastUploaded(context.Background(), &disk.QueryParams{"limit": "1"})
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.LastUploadedResourceList{}, t)
}

func TestMoveResource(t *testing.T) {

	vcr := useCassette("resources/move")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.Move(context.Background(), TEST_DIR_NAME_COPY, "test_dir_moved", nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestListPublicResources(t *testing.T) {

	vcr := useCassette("resources/get_public_res")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.ListPublic(context.Background(), &disk.QueryParams{"limit": "1"})
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.PublicResourcesList{}, t)
}

func TestPublishResource(t *testing.T) {

	vcr := useCassette("resources/publish")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.Publish(context.Background(), "test_dir_moved", nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestUnpublishResource(t *testing.T) {

	vcr := useCassette("resources/unpublish")
	defer vcr.Stop()
	resp, errorResponse := client.Resources.Unpublish(context.Background(), "test_dir_moved", nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestGetUploadLink(t *testing.T) {

	vcr := useCassette("resources/get_upload_link")
	defer vcr.Stop()

	resp, errorResponse := client.Resources.GetUploadLink(context.Background(), "upload_path")
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestUploadFile(t *testing.T) {

	vcr := useCassette("resources/upload_file")
	defer vcr.Stop()
	upload_link := "https://uploader7v.disk.yandex.net:443/upload-target/20221029T200308.792.utd.e8t7amr9zkrpoofffacoiggoz-k7v.6331006"

	errorResponse := client.Resources.Upload(context.Background(), "LICENSE", upload_link, nil)
	if errorResponse.StatusCode != http.StatusCreated {
		t.Fatalf("error: expect %v, got %v", 201, errorResponse.StatusCode)
	}
}

func TestDeleteResource(t *testing.T) {

	vcr := useCassette("resources/delete_resource")
	defer vcr.Stop()

	resp := client.Resources.Delete(context.Background(), TEST_DIR_NAME, false, nil)
	if resp != nil {
		t.Fatalf("error: expect %v, got %v", nil, resp)
	}
}
