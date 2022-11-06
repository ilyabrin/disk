package disk_test

import (
	"testing"

	"github.com/ilyabrin/disk"
)

func TestClient(t *testing.T) {
	client := disk.New("")
	if client != nil {
		t.Errorf("client should be %v, got %v", nil, client)
	}
}
