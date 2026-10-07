package disk

import (
	"context"
	"net/http"
	"testing"
)

// Real accounts report sizes and revisions far above 2^31, which must decode
// on 32-bit systems too, where int is 32 bits wide.

func TestDiskInfoDecodesLargeNumbers(t *testing.T) {
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"total_space": 1099511627776, "used_space": 87431285964,
			"trash_size": 5368709120, "max_file_size": 53687091200, "revision": 1602851010832695}`))
	})

	info, err := client.DiskInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.TotalSpace != 1099511627776 || info.UsedSpace != 87431285964 ||
		info.TrashSize != 5368709120 || info.MaxFileSize != 53687091200 || info.Revision != 1602851010832695 {
		t.Errorf("got %+v", info)
	}
}

func TestResourceDecodesAFileOverTwoGigabytes(t *testing.T) {
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name": "film.mkv", "path": "disk:/film.mkv", "type": "file",
			"size": 5368709120, "revision": 1602851010832695}`))
	})

	res, errResp := client.GetMetadata(context.Background(), "disk:/film.mkv")
	if errResp != nil {
		t.Fatal(errResp)
	}
	if res.Size != 5368709120 || res.Revision != 1602851010832695 {
		t.Errorf("size = %d, revision = %d", res.Size, res.Revision)
	}
}
