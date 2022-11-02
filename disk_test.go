package disk

import (
	"context"
	"reflect"
	"testing"
)

func TestDiskInfo(t *testing.T) {

	useCassette("disk/info")

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
