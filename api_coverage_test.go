package disk

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

// request is what a test server saw.
type request struct {
	method string
	path   string
	query  url.Values
}

// recordingClient answers every request with status and body, and records it.
func recordingClient(status int, body string) (*Client, *request) {
	var got request
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		got = request{r.Method, r.URL.Path, r.URL.Query()}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	return client, &got
}

func expectQuery(t *testing.T, got *request, method, path string, want map[string]string) {
	t.Helper()
	if got.method != method || got.path != path {
		t.Errorf("request = %s %s, want %s %s", got.method, got.path, method, path)
	}
	for k, v := range want {
		if got.query.Get(k) != v {
			t.Errorf("%s = %q, want %q (query %v)", k, got.query.Get(k), v, got.query)
		}
	}
}

const linkJSON = `{"href": "https://cloud-api.yandex.net/v1/disk/operations/abc", "method": "GET"}`

func TestCopyAndMoveSendTheirOptions(t *testing.T) {
	ctx := context.Background()
	opts := &CopyMoveOptions{Overwrite: true, ForceAsync: true}
	want := map[string]string{"from": "disk:/a", "path": "disk:/b", "overwrite": "true", "force_async": "true"}

	client, got := recordingClient(http.StatusAccepted, linkJSON)
	if _, errResp := client.CopyResourceWithOptions(ctx, "disk:/a", "disk:/b", opts); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodPost, "/v1/disk/resources/copy", want)

	client, got = recordingClient(http.StatusAccepted, linkJSON)
	if _, errResp := client.MoveResourceWithOptions(ctx, "disk:/a", "disk:/b", opts); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodPost, "/v1/disk/resources/move", want)

	// Without options nothing extra is sent.
	client, got = recordingClient(http.StatusCreated, linkJSON)
	if _, errResp := client.CopyResource(ctx, "disk:/a", "disk:/b"); errResp != nil {
		t.Fatal(errResp)
	}
	if got.query.Has("overwrite") || got.query.Has("force_async") {
		t.Errorf("plain copy sent options: %v", got.query)
	}
}

func TestDeleteSendsItsOptions(t *testing.T) {
	client, got := recordingClient(http.StatusAccepted, linkJSON)
	err := client.DeleteResourceWithOptions(context.Background(), "disk:/a.txt",
		&DeleteOptions{Permanently: true, ForceAsync: true, MD5: "abc123"})
	if err != nil {
		t.Fatal(err)
	}
	expectQuery(t, got, http.MethodDelete, "/v1/disk/resources",
		map[string]string{"path": "disk:/a.txt", "permanently": "true", "force_async": "true", "md5": "abc123"})
}

func TestListsSendTheirFilters(t *testing.T) {
	ctx := context.Background()

	client, got := recordingClient(http.StatusOK, `{"items": [], "limit": 5}`)
	if _, errResp := client.GetLastUploadedResourcesWithOptions(ctx, &PaginationOptions{Limit: 5},
		&FilesOptions{MediaType: []string{"image", "video"}, PreviewSize: "M", Sort: "name"}); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodGet, "/v1/disk/resources/last-uploaded",
		map[string]string{"limit": "5", "media_type": "image,video", "preview_size": "M"})
	if got.query.Has("sort") || got.query.Has("offset") {
		t.Errorf("last-uploaded takes neither sort nor offset: %v", got.query)
	}

	client, got = recordingClient(http.StatusOK, `{"items": [], "limit": 20, "offset": 0}`)
	if _, errResp := client.GetPublicResourcesWithOptions(ctx, nil,
		&PublicResourcesOptions{Type: "dir", PreviewCrop: true}); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodGet, "/v1/disk/resources/public",
		map[string]string{"type": "dir", "preview_crop": "true"})

	client, got = recordingClient(http.StatusOK, `{"path": "trash:/", "type": "dir", "_embedded": {"items": []}}`)
	if _, err := client.ListTrashResourcesWithOptions(ctx, "", &TrashListOptions{Limit: 10, Sort: "-deleted"}); err != nil {
		t.Fatal(err)
	}
	expectQuery(t, got, http.MethodGet, "/v1/disk/trash/resources", map[string]string{"limit": "10", "sort": "-deleted"})
}

func TestRestoreSaveAndUploadByURLSendTheirOptions(t *testing.T) {
	ctx := context.Background()

	client, got := recordingClient(http.StatusAccepted, linkJSON)
	if _, err := client.RestoreFromTrashWithOptions(ctx, "trash:/a.txt",
		&RestoreOptions{Overwrite: true, Name: "b.txt", ForceAsync: true}); err != nil {
		t.Fatal(err)
	}
	expectQuery(t, got, http.MethodPut, "/v1/disk/trash/resources/restore",
		map[string]string{"overwrite": "true", "name": "b.txt", "force_async": "true"})

	client, got = recordingClient(http.StatusAccepted, linkJSON)
	if _, errResp := client.SavePublicResourceWithOptions(ctx, "key",
		&SavePublicResourceOptions{SavePath: "disk:/Inbox", ForceAsync: true}); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodPost, "/v1/disk/public/resources/save-to-disk",
		map[string]string{"public_key": "key", "save_path": "disk:/Inbox", "force_async": "true"})

	client, got = recordingClient(http.StatusAccepted, linkJSON)
	if _, errResp := client.UploadFileWithOptions(ctx, "disk:/f.zip", "https://example.com/f.zip",
		&UploadFromURLOptions{DisableRedirects: true}); errResp != nil {
		t.Fatal(errResp)
	}
	expectQuery(t, got, http.MethodPost, "/v1/disk/resources/upload",
		map[string]string{"url": "https://example.com/f.zip", "disable_redirects": "true"})
}

func TestNewFieldsDecode(t *testing.T) {
	client, _ := recordingClient(http.StatusOK, `{
		"total_space": 2199023255552, "used_space": 1000, "paid_max_file_size": 53687091200,
		"mail_size": 300, "disk_size": 700, "photounlim_size": 0, "will_be_overdrawn": true,
		"monthly_traffic_limit": 1099511627776, "deletion_restriction_days": 90,
		"reg_time": "2012-04-04T12:00:00+00:00",
		"file_size_limit_upgrades": {"paid": 53687091200, "pro": 107374182400},
		"monthly_traffic_limit_upgrades": {"pro": 2199023255552},
		"system_folders": {"scans": "disk:/Сканы/", "attach": "disk:/Почтовые вложения/",
			"messenger": "disk:/Файлы Мессенджера/", "calendar": "disk:/Материалы встреч/"},
		"user": {"login": "demo", "is_child": true, "reg_time": "2012-04-04T12:00:00+00:00"}
	}`)
	d, err := client.DiskInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if d.PaidMaxFileSize != 53687091200 || d.MailSize != 300 || d.DiskSize != 700 || !d.WillBeOverdrawn ||
		d.MonthlyTrafficLimit != 1099511627776 || d.DeletionRestrictionDays != 90 || d.RegTime == "" {
		t.Errorf("disk fields: %+v", d)
	}
	if d.FileSizeLimitUpgrades == nil || d.FileSizeLimitUpgrades.Pro != 107374182400 ||
		d.MonthlyTrafficLimitUpgrades == nil || d.MonthlyTrafficLimitUpgrades.Pro != 2199023255552 {
		t.Errorf("upgrades: %+v %+v", d.FileSizeLimitUpgrades, d.MonthlyTrafficLimitUpgrades)
	}
	if f := d.SystemFolders; f == nil || f.Scans == "" || f.Attach == "" || f.Messenger == "" || f.Calendar == "" {
		t.Errorf("system folders: %+v", f)
	}
	if !d.User.IsChild || d.User.RegTime == "" {
		t.Errorf("user: %+v", d.User)
	}

	client, _ = recordingClient(http.StatusOK, `{"name": "cat.jpg", "path": "disk:/cat.jpg", "type": "file",
		"exif": {"date_time": "2026-08-16T23:47:00+00:00", "gps_latitude": 64.1466, "gps_longitude": -21.9426},
		"sizes": [{"name": "ORIGINAL", "url": "https://downloader.disk.yandex.ru/x"}, {"name": "S", "url": "https://downloader.disk.yandex.ru/s"}]}`)
	r, errResp := client.GetMetadata(context.Background(), "disk:/cat.jpg")
	if errResp != nil {
		t.Fatal(errResp)
	}
	if r.Exif == nil || r.Exif.Latitude != 64.1466 || r.Exif.Longitude != -21.9426 {
		t.Errorf("exif: %+v", r.Exif)
	}
	if len(r.Sizes) != 2 || r.Sizes[1].Name != "S" || r.Sizes[1].URL == "" {
		t.Errorf("sizes: %+v", r.Sizes)
	}
}

func TestCoordinateToleratesOddShapes(t *testing.T) {
	cases := map[string]Coordinate{`64.5`: 64.5, `"-21.25"`: -21.25, `{}`: 0, `null`: 0, `"n/a"`: 0}
	for in, want := range cases {
		var c Coordinate
		if err := c.UnmarshalJSON([]byte(in)); err != nil || c != want {
			t.Errorf("%s -> %v, %v; want %v", in, c, err, want)
		}
	}
}
