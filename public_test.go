package disk_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestGetMetadataForPublicResource(t *testing.T) {

	UseCassette("/public/get_meta")

	resp, errorResponse := client.GetMetadataForPublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	publicResource := new(disk.PublicResource)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(publicResource).Kind() {
		t.Fatalf("error: expect %v, got %v", publicResource, resp)
	}
}

func TestGetDownloadURLForPublicResource(t *testing.T) {

	UseCassette("/public/download_url")

	resp, errorResponse := client.GetDownloadURLForPublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestSavePublicResource(t *testing.T) {

	UseCassette("/public/save")

	resp, errorResponse := client.SavePublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(disk.Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}
