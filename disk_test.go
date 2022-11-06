package disk_test

import (
	"context"
	"testing"

	"github.com/ilyabrin/disk"
)

func TestDiskInfo(t *testing.T) {

	vcr := useCassette("disk/info")
	defer vcr.Stop()

	resp, errorResponse := client.Disk.Info(context.Background(), nil)
	if errorResponse != nil {
		t.Fatal("errorResponse should be nil")
	}

	disk := new(disk.Disk)

	checkTypes(resp, disk, t)

	if client.ReqURL() != client.ApiURL() {
		t.Fatalf("error: expect %v, got %v", client.ReqURL(), client.ApiURL())
	}
}
