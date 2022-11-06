package disk_test

import (
	"context"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestPublicMeta(t *testing.T) {

	vcr := useCassette("public/get_meta")
	defer vcr.Stop()

	resp, errorResponse := client.Public.Meta(context.Background(), TEST_PUBLIC_RESOURCE, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.PublicResource{}, t)
}

func TestPublicDownloadURL(t *testing.T) {

	vcr := useCassette("public/download_url")
	defer vcr.Stop()

	resp, errorResponse := client.Public.DownloadURL(context.Background(), TEST_PUBLIC_RESOURCE, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}

func TestPublicSave(t *testing.T) {

	vcr := useCassette("public/save")
	defer vcr.Stop()

	resp, errorResponse := client.Public.Save(context.Background(), TEST_PUBLIC_RESOURCE, nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	checkTypes(resp, &disk.Link{}, t)
}
