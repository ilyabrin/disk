package disk_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestDiskInfo(t *testing.T) {

	UseCassette("disk/info")

	resp, errorResponse := client.DiskInfo(context.Background(), nil)

	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	disk := new(disk.Disk)

	if reflect.TypeOf(resp).Kind() != reflect.TypeOf(disk).Kind() {
		t.Fatalf("error: expect %v, got %v", disk, resp)
	}

	if client.ReqURL() != client.ApiURL() {
		t.Fatalf("error: expect %v, got %v", client.ReqURL(), client.ApiURL())
	}
}
