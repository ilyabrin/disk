package disk

import (
	"context"
	"net/http"
	"testing"
)

// Bodies of known length must go with Content-Length: sent chunked, the API
// read PATCH and PUT bodies as empty.
func TestJSONBodiesCarryContentLength(t *testing.T) {
	var length int64
	var chunked bool
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		length = r.ContentLength
		chunked = len(r.TransferEncoding) > 0
		_, _ = w.Write([]byte(`{"name": "a", "path": "disk:/a", "type": "file"}`))
	})

	props := map[string]map[string]string{"custom_properties": {"k": "v"}}
	if _, errResp := client.UpdateMetadata(context.Background(), "disk:/a", props); errResp != nil {
		t.Fatal(errResp)
	}
	if length <= 0 || chunked {
		t.Errorf("Content-Length = %d, chunked = %v; want a known length", length, chunked)
	}
}
