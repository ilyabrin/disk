package disk

import (
	"context"
	"reflect"
	"testing"
)

const TEST_PUBLIC_RESOURCE = "https://disk.yandex.ru/d/tCgV7GyS3QAYvg"

func TestGetMetadataForPublicResource(t *testing.T) {

	useCassette("/public/get_meta")

	resp, errorResponse := client.GetMetadataForPublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	publicResource := new(PublicResource)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(publicResource).Kind() {
		t.Fatalf("error: expect %v, got %v", publicResource, resp)
	}
}

func TestGetDownloadURLForPublicResource(t *testing.T) {

	useCassette("/public/download_url")

	resp, errorResponse := client.GetDownloadURLForPublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}

func TestSavePublicResource(t *testing.T) {

	useCassette("/public/save")

	resp, errorResponse := client.SavePublicResource(context.Background(), TEST_PUBLIC_RESOURCE, nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	link := new(Link)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(link).Kind() {
		t.Fatalf("error: expect %v, got %v", link, resp)
	}
}
