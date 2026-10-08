package disk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

// settingsServer records the requests of a publish-with-settings flow and
// answers the PATCH with patchStatus.
type settingsServer struct {
	mu       sync.Mutex
	requests []string
	patch    map[string]any
}

func (s *settingsServer) client(patchStatus int) *Client {
	return mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.requests = append(s.requests, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPatch:
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &s.patch)
			w.WriteHeader(patchStatus)
			if patchStatus != http.StatusOK {
				_, _ = w.Write([]byte(`{"error": "FieldValidationError", "description": "Bad."}`))
			}
		default:
			_, _ = w.Write([]byte(`{"href": "https://cloud-api.yandex.net/v1/disk/resources?path=disk%3A%2Fa", "method": "GET"}`))
		}
	})
}

func TestPublishWithSettingsPublishesThenPatches(t *testing.T) {
	var srv settingsServer
	expires := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	_, errResp := srv.client(http.StatusOK).PublishResourceWithSettings(context.Background(), "disk:/a",
		&PublicSettings{Password: "secret", ExpiresAt: expires})
	if errResp != nil {
		t.Fatal(errResp)
	}
	want := []string{"PUT /v1/disk/resources/publish", "PATCH /v1/disk/public/resources/public-settings"}
	if len(srv.requests) != 2 || srv.requests[0] != want[0] || srv.requests[1] != want[1] {
		t.Errorf("requests = %v, want %v", srv.requests, want)
	}
	// The API takes the expiry as a Unix time, not as a duration.
	if srv.patch["password"] != "secret" || srv.patch["available_until"] != float64(expires.Unix()) {
		t.Errorf("patch body = %v", srv.patch)
	}
	if _, ok := srv.patch["accesses"]; ok {
		t.Error("accesses must not be sent unless asked for: they close the link on personal accounts")
	}
}

func TestPublishWithSettingsUnpublishesWhenTheyFail(t *testing.T) {
	var srv settingsServer
	_, errResp := srv.client(http.StatusBadRequest).PublishResourceWithSettings(context.Background(), "disk:/a",
		&PublicSettings{Password: "secret"})
	if errResp == nil {
		t.Fatal("a failed PATCH must be reported")
	}
	last := srv.requests[len(srv.requests)-1]
	if last != "PUT /v1/disk/resources/unpublish" {
		t.Errorf("requests = %v; the link must not stay open without its protection", srv.requests)
	}
}

func TestPublishWithEmptySettingsIsAPlainPublish(t *testing.T) {
	var srv settingsServer
	if _, errResp := srv.client(http.StatusOK).PublishResourceWithSettings(context.Background(), "disk:/a", &PublicSettings{}); errResp != nil {
		t.Fatal(errResp)
	}
	if len(srv.requests) != 1 {
		t.Errorf("requests = %v, want only the publish", srv.requests)
	}
}

func TestAccessesGetTheirType(t *testing.T) {
	body, err := (&PublicSettings{Accesses: []PublicAccess{{Macros: []string{MacroEmployees}, OrgID: 7, Rights: []string{RightRead}}}}).body()
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Accesses []map[string]any `json:"accesses"`
	}
	_ = json.Unmarshal(body, &got)
	if len(got.Accesses) != 1 || got.Accesses[0]["type"] != "macro" {
		t.Errorf("body = %s; the API rejects an access without a type", body)
	}
}

func TestUpdatePublicSettingsRejectsEmptySettings(t *testing.T) {
	var srv settingsServer
	if errResp := srv.client(http.StatusOK).UpdatePublicSettings(context.Background(), "disk:/a", &PublicSettings{}); errResp == nil {
		t.Error("empty settings must be refused before a request is made")
	}
	if len(srv.requests) != 0 {
		t.Errorf("requests = %v", srv.requests)
	}
}

func TestGetPublicSettingsReadsTheExpiry(t *testing.T) {
	for body, want := range map[string]int64{`{"available_until": 1791419390}`: 1791419390, `{"available_until": null}`: 0} {
		client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(body))
		})
		info, errResp := client.GetPublicSettings(context.Background(), "disk:/a")
		if errResp != nil {
			t.Fatal(errResp)
		}
		got := int64(0)
		if !info.ExpiresAt.IsZero() {
			got = info.ExpiresAt.Unix()
		}
		if got != want {
			t.Errorf("%s: expires %d, want %d", body, got, want)
		}
	}
}
