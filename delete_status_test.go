package disk

import (
	"context"
	"net/http"
	"testing"
)

// The real API answers a delete with 204, or 202 for a large folder deleted
// in the background. Both mean success.
func TestDeleteResourceAcceptsTheRealStatusCodes(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusAccepted, http.StatusOK} {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("method = %s, want DELETE", r.Method)
			}
			w.WriteHeader(status)
			if status == http.StatusAccepted {
				_, _ = w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/operations/abc", "method": "GET"}`))
			}
		})
		if err := client.DeleteResource(context.Background(), "disk:/old.txt", false); err != nil {
			t.Errorf("status %d: %v", status, err)
		}
	}
}

func TestDeleteResourceReportsAMissingResource(t *testing.T) {
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "DiskNotFoundError", "description": "Resource not found."}`))
	})
	if err := client.DeleteResource(context.Background(), "disk:/missing.txt", false); err == nil {
		t.Error("a 404 must be reported as an error")
	}
}
