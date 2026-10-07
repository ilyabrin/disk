package disk

import (
	"context"
	"net/http"
	"testing"
)

// yandexTrashResponse is shaped like the real answer to
// GET /v1/disk/trash/resources: the trash root is itself a resource, and its
// contents sit under "_embedded", exactly as for an ordinary folder.
const yandexTrashResponse = `{
	"path": "trash:/",
	"name": "trash",
	"type": "dir",
	"created": "2024-01-01T00:00:00+00:00",
	"modified": "2026-09-30T10:00:00+00:00",
	"_embedded": {
		"path": "trash:/",
		"sort": "",
		"limit": 100,
		"offset": 0,
		"total": 2,
		"items": [
			{
				"path": "trash:/report.pdf_1727690400",
				"name": "report.pdf",
				"type": "file",
				"size": 20480,
				"origin_path": "disk:/Documents/report.pdf",
				"deleted": "2026-09-30T10:00:00+00:00",
				"created": "2026-09-01T09:00:00+00:00",
				"modified": "2026-09-01T09:00:00+00:00"
			},
			{
				"path": "trash:/Old photos_1727690500",
				"name": "Old photos",
				"type": "dir",
				"origin_path": "disk:/Old photos",
				"deleted": "2026-09-30T10:01:40+00:00",
				"created": "2025-05-01T09:00:00+00:00",
				"modified": "2025-05-01T09:00:00+00:00"
			}
		]
	}
}`

func TestListTrashResourcesReadsTheRealResponseShape(t *testing.T) {
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(yandexTrashResponse))
	})

	list, err := client.ListTrashResources(context.Background(), "", 100, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(list.Items) != 2 {
		t.Fatalf("got %d items, want the 2 in _embedded", len(list.Items))
	}
	first := list.Items[0]
	if first.Name != "report.pdf" || first.OriginPath != "disk:/Documents/report.pdf" {
		t.Errorf("first item = %q from %q", first.Name, first.OriginPath)
	}
	if first.Path != "trash:/report.pdf_1727690400" {
		t.Errorf("path = %q; restore and delete need the trash path", first.Path)
	}
	if list.Items[1].Type != "dir" {
		t.Errorf("second item type = %q, want dir", list.Items[1].Type)
	}
	if list.Total != 2 || list.Limit != 100 || list.Path != "trash:/" {
		t.Errorf("total=%d limit=%d path=%q, want 2, 100, trash:/", list.Total, list.Limit, list.Path)
	}
}

func TestListTrashResourcesHandlesAnEmptyTrash(t *testing.T) {
	client := mockedHttpClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path": "trash:/", "name": "trash", "type": "dir",
			"_embedded": {"path": "trash:/", "items": [], "limit": 100, "offset": 0, "total": 0}}`))
	})

	list, err := client.ListTrashResources(context.Background(), "", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 0 || list.Total != 0 {
		t.Errorf("got %d items, total %d; want an empty list", len(list.Items), list.Total)
	}
}
