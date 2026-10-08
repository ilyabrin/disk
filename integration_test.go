//go:build integration

// Integration tests against the real Yandex.Disk API. They do not run with a
// plain `go test`; run them with a token of an account you can spare:
//
//	YANDEX_DISK_ACCESS_TOKEN=... go test -tags integration -run Integration -v ./
//
// Everything they create lives in one new folder, disk:/disk-it-<time>, which
// is deleted at the end, permanently. A guard in the HTTP transport refuses
// any request that would change something outside that folder, or any trash
// item the tests did not put there, and never lets the whole trash be
// emptied. Calls that read the whole Disk (disk info, file lists) only log
// counts, never names.
package disk

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// guard is an http.RoundTripper that lets a request through only if every
// path it changes is inside root, or is a trash item the tests created.
type guard struct {
	base  http.RoundTripper
	root  string // "disk:/disk-it-..."
	mu    sync.Mutex
	trash map[string]bool // trash paths whose origin is inside root
}

func (g *guard) allowTrash(path string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.trash[path] = true
}

func (g *guard) inside(p string) bool {
	if !strings.HasPrefix(p, "disk:") {
		p = "disk:" + p
	}
	return p == g.root || strings.HasPrefix(p, g.root+"/")
}

func (g *guard) RoundTrip(r *http.Request) (*http.Response, error) {
	// Reads cannot change anything, and uploads go to the storage host with
	// a link the API issued for a path the guard already checked.
	if r.Method == http.MethodGet || r.URL.Host != "cloud-api.yandex.net" {
		return g.base.RoundTrip(r)
	}
	q := r.URL.Query()
	isTrash := strings.Contains(r.URL.Path, "/trash/")
	if isTrash && q.Get("path") == "" {
		return nil, fmt.Errorf("guard: refusing %s %s without a path: it would touch the whole trash", r.Method, r.URL.Path)
	}
	for _, key := range []string{"path", "from", "save_path"} {
		v := q.Get(key)
		if v == "" {
			continue
		}
		if strings.HasPrefix(v, "trash:") {
			g.mu.Lock()
			ok := g.trash[v]
			g.mu.Unlock()
			if !ok {
				return nil, fmt.Errorf("guard: refusing %s on trash item %s the tests did not create", r.Method, v)
			}
			continue
		}
		if !g.inside(v) {
			return nil, fmt.Errorf("guard: refusing %s with %s=%s outside %s", r.Method, key, v, g.root)
		}
	}
	return g.base.RoundTrip(r)
}

type sandbox struct {
	t     *testing.T
	c     *Client
	g     *guard
	root  string
	ctx   context.Context
	local string // local temp dir
}

func (s *sandbox) p(name string) string { return s.root + "/" + name }

// waitOperation follows a background operation behind link until it ends.
func (s *sandbox) waitOperation(link *Link) {
	s.t.Helper()
	if link == nil || !strings.Contains(link.Href, "/operations/") {
		return // finished at once
	}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		op, err := s.c.GetOperationStatus(s.ctx, link.Href)
		if err != nil {
			s.t.Fatalf("operation status: %v", err)
		}
		switch op.Status {
		case "success":
			return
		case "failed":
			s.t.Fatalf("operation %s failed", link.Href)
		}
		time.Sleep(time.Second)
	}
	s.t.Fatalf("operation %s did not finish in time", link.Href)
}

// waitFor polls cond until it holds, for things the API applies eventually.
func (s *sandbox) waitFor(what string, cond func() bool) {
	s.t.Helper()
	for i := 0; i < 30; i++ {
		if cond() {
			return
		}
		time.Sleep(time.Second)
	}
	s.t.Fatalf("timed out waiting for %s", what)
}

func (s *sandbox) file(name, content string) string {
	s.t.Helper()
	p := filepath.Join(s.local, strings.ReplaceAll(name, "/", "_"))
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		s.t.Fatal(err)
	}
	return p
}

func (s *sandbox) upload(name, content string) {
	s.t.Helper()
	if _, err := s.c.UploadFileFromPath(s.ctx, s.file(name, content), s.p(name), &UploadOptions{Overwrite: true}); err != nil {
		s.t.Fatalf("upload %s: %v", name, err)
	}
}

func (s *sandbox) exists(name string) bool {
	_, errResp := s.c.GetMetadata(s.ctx, s.p(name))
	return errResp == nil
}

// trashItem finds the trash entry deleted from s.p(name) and lets the guard
// touch it.
func (s *sandbox) trashItem(name string) string {
	s.t.Helper()
	origin := s.p(name)
	var found string
	s.waitFor("trash entry of "+name, func() bool {
		list, err := s.c.ListTrashResourcesWithOptions(s.ctx, "", &TrashListOptions{Limit: 100, Sort: "-deleted"})
		if err != nil {
			s.t.Fatalf("list trash: %v", err)
		}
		for _, item := range list.Items {
			if item.OriginPath == origin {
				found = item.Path
				return true
			}
		}
		return false
	})
	s.g.allowTrash(found)
	return found
}

func TestIntegration(t *testing.T) {
	token := os.Getenv("YANDEX_DISK_ACCESS_TOKEN")
	if token == "" {
		t.Skip("set YANDEX_DISK_ACCESS_TOKEN to run against the real API")
	}
	client, err := New(token)
	if err != nil {
		t.Fatal(err)
	}
	client.Logger.SetLevel(SILENT)

	root := "disk:/disk-it-" + time.Now().Format("20060102-150405")
	g := &guard{base: client.HTTPClient.Transport, root: root, trash: map[string]bool{}}
	if g.base == nil {
		g.base = http.DefaultTransport
	}
	client.HTTPClient.Transport = g

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	s := &sandbox{t: t, c: client, g: g, root: root, ctx: ctx, local: t.TempDir()}
	t.Logf("sandbox: %s", root)

	if _, errResp := client.CreateDir(ctx, root); errResp != nil {
		t.Fatalf("create sandbox: %+v", errResp)
	}
	t.Cleanup(func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer ccancel()
		if err := client.DeleteResourceWithOptions(cctx, root, &DeleteOptions{Permanently: true}); err != nil {
			t.Errorf("CLEANUP FAILED, delete %s by hand: %v", root, err)
			return
		}
		t.Logf("sandbox deleted permanently: %s", root)
	})

	t.Run("read-only overview", func(t *testing.T) {
		s.t = t
		info, err := client.DiskInfo(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if info.TotalSpace <= 0 || info.User == nil {
			t.Errorf("disk info looks empty: total=%d", info.TotalSpace)
		}
		t.Logf("disk: %s of %s used, paid=%v, max file %s",
			FormatFileSize(info.UsedSpace), FormatFileSize(info.TotalSpace), info.IsPaid, FormatFileSize(info.MaxFileSize))
		if files, errResp := client.GetSortedFilesWithPagination(ctx, &PaginationOptions{Limit: 3}); errResp != nil {
			t.Errorf("flat file list: %+v", errResp)
		} else {
			t.Logf("flat file list: %d items on the first page", len(files.Items))
		}
		if last, errResp := client.GetLastUploadedResourcesWithOptions(ctx, &PaginationOptions{Limit: 3}, &FilesOptions{MediaType: []string{"image"}}); errResp != nil {
			t.Errorf("last uploaded: %+v", errResp)
		} else {
			t.Logf("last uploaded images: %d", len(last.Items))
		}
		if pub, errResp := client.GetPublicResourcesWithOptions(ctx, &PaginationOptions{Limit: 3}, nil); errResp != nil {
			t.Errorf("public resources: %+v", errResp)
		} else {
			t.Logf("public resources: %d on the first page", len(pub.Items))
		}
		if trash, err := client.ListTrashResourcesWithOptions(ctx, "", &TrashListOptions{Limit: 3}); err != nil {
			t.Errorf("trash: %v", err)
		} else {
			t.Logf("trash: %d items in total", trash.Total)
		}
	})

	t.Run("folders", func(t *testing.T) {
		s.t = t
		if errResp := client.CreateDirAll(ctx, s.p("a/b/c")); errResp != nil {
			t.Fatalf("%+v", errResp)
		}
		res, errResp := client.GetMetadata(ctx, s.p("a/b/c"))
		if errResp != nil || res.Type != "dir" {
			t.Fatalf("metadata: %+v %+v", res, errResp)
		}
		if errResp := client.CreateDirAll(ctx, s.p("a/b/c")); errResp != nil {
			t.Errorf("CreateDirAll on an existing path must succeed: %+v", errResp)
		}
	})

	t.Run("upload, overwrite and download", func(t *testing.T) {
		s.t = t
		s.upload("hello.txt", "hello, disk")
		res, errResp := client.GetMetadata(ctx, s.p("hello.txt"))
		if errResp != nil || res.Size != int64(len("hello, disk")) {
			t.Fatalf("metadata: %+v %+v", res, errResp)
		}

		// Uploading over it without overwrite must be refused...
		_, err := client.UploadFileFromPath(ctx, s.file("v2.txt", "second version"), s.p("hello.txt"), &UploadOptions{})
		if err == nil {
			t.Error("upload over an existing file without overwrite should fail")
		}
		// ...and with overwrite must replace it (the 409 bug of v1.3).
		if _, err := client.UploadFileFromPath(ctx, s.file("v2.txt", "second version"), s.p("hello.txt"), &UploadOptions{Overwrite: true}); err != nil {
			t.Fatalf("overwrite upload: %v", err)
		}

		dst := filepath.Join(s.local, "downloaded.txt")
		if err := client.DownloadFileToPath(ctx, s.p("hello.txt"), dst, &DownloadOptions{Overwrite: true}); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "second version" {
			t.Errorf("downloaded %q, want the overwritten content", got)
		}
		if link, errResp := client.GetDownloadURL(ctx, s.p("hello.txt")); errResp != nil || link.Href == "" {
			t.Errorf("download URL: %+v %+v", link, errResp)
		}
	})

	t.Run("large upload in chunks", func(t *testing.T) {
		s.t = t
		data := bytes.Repeat([]byte("0123456789abcdef"), 6*1024*1024/16) // 6 MB
		local := filepath.Join(s.local, "big.bin")
		if err := os.WriteFile(local, data, 0o600); err != nil {
			t.Fatal(err)
		}
		res, err := client.UploadLargeFileFromPath(ctx, local, s.p("big.bin"), 2, nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.Size != int64(len(data)) {
			t.Errorf("size = %d, want %d", res.Size, len(data))
		}
	})

	t.Run("custom properties", func(t *testing.T) {
		s.t = t
		props := map[string]map[string]string{"custom_properties": {"suite": "disk-it"}}
		if _, errResp := client.UpdateMetadata(ctx, s.p("hello.txt"), props); errResp != nil {
			t.Fatalf("%+v", errResp)
		}
		res, errResp := client.GetMetadataWithOptions(ctx, s.p("hello.txt"), &ResourceOptions{Fields: []string{"custom_properties"}})
		if errResp != nil {
			t.Fatal(errResp)
		}
		if m, _ := res.CustomProperties.(map[string]any); m["suite"] != "disk-it" {
			t.Errorf("custom properties = %v", res.CustomProperties)
		}
	})

	t.Run("listing pages", func(t *testing.T) {
		s.t = t
		if errResp := client.CreateDirAll(ctx, s.p("list")); errResp != nil {
			t.Fatal(errResp)
		}
		for i := 0; i < 7; i++ {
			s.upload(fmt.Sprintf("list/f%02d.txt", i), "x")
		}
		var names []string
		for offset := 0; ; offset += 3 {
			page, errResp := client.GetMetadataWithOptions(ctx, s.p("list"), &ResourceOptions{Limit: 3, Offset: offset, Sort: "name"})
			if errResp != nil {
				t.Fatal(errResp)
			}
			for _, item := range page.Embedded.Items {
				names = append(names, item.Name)
			}
			if offset+3 >= page.Embedded.Total {
				break
			}
		}
		if len(names) != 7 || names[0] != "f00.txt" || names[6] != "f06.txt" {
			t.Errorf("pages gave %v", names)
		}
	})

	t.Run("copy, move and background operations", func(t *testing.T) {
		s.t = t
		link, errResp := client.CopyResource(ctx, s.p("hello.txt"), s.p("copy.txt"))
		if errResp != nil {
			t.Fatal(errResp)
		}
		s.waitOperation(link)
		if _, errResp := client.CopyResource(ctx, s.p("hello.txt"), s.p("copy.txt")); errResp == nil {
			t.Error("copy onto an existing file without overwrite should fail")
		}
		link, errResp = client.CopyResourceWithOptions(ctx, s.p("hello.txt"), s.p("copy.txt"), &CopyMoveOptions{Overwrite: true})
		if errResp != nil {
			t.Fatalf("copy with overwrite: %+v", errResp)
		}
		s.waitOperation(link)

		link, errResp = client.MoveResource(ctx, s.p("copy.txt"), s.p("renamed.txt"))
		if errResp != nil {
			t.Fatal(errResp)
		}
		s.waitOperation(link)
		if s.exists("copy.txt") || !s.exists("renamed.txt") {
			t.Error("move did not rename")
		}

		// A forced background copy of a folder exercises GetOperationStatus.
		link, errResp = client.CopyResourceWithOptions(ctx, s.p("list"), s.p("list-copy"), &CopyMoveOptions{ForceAsync: true})
		if errResp != nil {
			t.Fatalf("async copy: %+v", errResp)
		}
		if !strings.Contains(link.Href, "/operations/") {
			t.Errorf("force_async did not return an operation: %s", link.Href)
		}
		s.waitOperation(link)
		if !s.exists("list-copy/f06.txt") {
			t.Error("async copy is missing files")
		}
	})

	t.Run("batch operations", func(t *testing.T) {
		s.t = t
		ops := map[string]string{s.p("list/f00.txt"): s.p("batch-0.txt"), s.p("list/f01.txt"): s.p("batch-1.txt")}
		status, err := client.BatchCopyFiles(ctx, ops, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.WaitForBatchOperation(ctx, status, time.Second); err != nil {
			t.Fatal(err)
		}
		if status.Failed != 0 || !s.exists("batch-0.txt") || !s.exists("batch-1.txt") {
			t.Errorf("batch copy: %+v", status.GetSummary())
		}
		del, err := client.BatchDeleteFilesSimple(ctx, []string{s.p("batch-0.txt"), s.p("batch-1.txt")}, true)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.WaitForBatchOperation(ctx, del, time.Second); err != nil {
			t.Fatal(err)
		}
		if del.Failed != 0 {
			t.Errorf("batch delete: %+v", del.GetSummary())
		}
	})

	t.Run("publish and the public API", func(t *testing.T) {
		s.t = t
		if _, errResp := client.PublishResource(ctx, s.p("hello.txt")); errResp != nil {
			t.Fatal(errResp)
		}
		res, errResp := client.GetMetadata(ctx, s.p("hello.txt"))
		if errResp != nil || res.PublicURL == "" || res.PublicKey == "" {
			t.Fatalf("no public link: %+v", errResp)
		}
		pub, errResp := client.GetMetadataForPublicResource(ctx, res.PublicKey)
		if errResp != nil || pub.Name != "hello.txt" {
			t.Errorf("public metadata: %+v %+v", pub, errResp)
		}
		if link, errResp := client.GetDownloadURLForPublicResource(ctx, res.PublicKey); errResp != nil || link.Href == "" {
			t.Errorf("public download URL: %+v", errResp)
		}
		if errResp := client.CreateDirAll(ctx, s.p("saved")); errResp != nil {
			t.Fatal(errResp)
		}
		link, errResp := client.SavePublicResourceWithOptions(ctx, res.PublicKey,
			&SavePublicResourceOptions{SavePath: s.p("saved"), Name: "from-public.txt"})
		if errResp != nil {
			t.Errorf("save public resource into the sandbox: %+v", errResp)
		} else {
			s.waitOperation(link)
			if !s.exists("saved/from-public.txt") {
				t.Error("saved public resource is not where save_path said")
			}
		}
		if _, errResp := client.UnpublishResource(ctx, s.p("hello.txt")); errResp != nil {
			t.Fatal(errResp)
		}
		res, _ = client.GetMetadata(ctx, s.p("hello.txt"))
		if res != nil && res.PublicURL != "" {
			t.Error("still published after unpublish")
		}
	})

	// Checked as a stranger would see the link: no token. The public API
	// answers 403 for a link behind a password and 404 for an expired one.
	t.Run("public link settings", func(t *testing.T) {
		s.t = t
		s.upload("settings.txt", "link settings")
		path := s.p("settings.txt")
		stranger := func() int {
			res, errResp := client.GetMetadata(ctx, path)
			if errResp != nil || res.PublicKey == "" {
				t.Fatalf("not published: %+v", errResp)
			}
			u := "https://cloud-api.yandex.net/v1/disk/public/resources?public_key=" + url.QueryEscape(res.PublicKey)
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("stranger request: %v", err)
			}
			_ = resp.Body.Close()
			return resp.StatusCode
		}

		if _, errResp := client.PublishResourceWithSettings(ctx, path, &PublicSettings{Password: "Disk-it-1234"}); errResp != nil {
			t.Fatalf("publish with a password: %+v", errResp)
		}
		if code := stranger(); code != http.StatusForbidden {
			t.Errorf("a stranger opened a password-protected link: %d", code)
		}
		_, _ = client.UnpublishResource(ctx, path)

		expires := time.Now().Add(20 * time.Second)
		if _, errResp := client.PublishResourceWithSettings(ctx, path, &PublicSettings{ExpiresAt: expires}); errResp != nil {
			t.Fatalf("publish with an expiry: %+v", errResp)
		}
		info, errResp := client.GetPublicSettings(ctx, path)
		if errResp != nil || info.ExpiresAt.Unix() != expires.Unix() {
			t.Errorf("stored expiry = %+v %+v, want %v", info, errResp, expires.Unix())
		}
		if code := stranger(); code != http.StatusOK {
			t.Errorf("the link is closed before it expires: %d", code)
		}
		time.Sleep(time.Until(expires) + 15*time.Second)
		if code := stranger(); code != http.StatusNotFound {
			t.Errorf("a stranger opened an expired link: %d", code)
		}

		_, _ = client.UnpublishResource(ctx, path)
		if _, errResp := client.PublishResource(ctx, path); errResp != nil {
			t.Fatal(errResp)
		}
		if errResp := client.UpdatePublicSettings(ctx, path, &PublicSettings{Password: "Disk-it-5678"}); errResp != nil {
			t.Fatalf("add a password to an open link: %+v", errResp)
		}
		if code := stranger(); code != http.StatusForbidden {
			t.Errorf("a stranger opened a link after a password was added: %d", code)
		}
		_, _ = client.UnpublishResource(ctx, path)
	})

	t.Run("upload from a URL", func(t *testing.T) {
		s.t = t
		link, errResp := client.UploadFile(ctx, s.p("license.txt"), "https://raw.githubusercontent.com/ilyabrin/disk/main/LICENSE")
		if errResp != nil {
			t.Fatal(errResp)
		}
		s.waitOperation(link)
		s.waitFor("license.txt to arrive", func() bool { return s.exists("license.txt") })
	})

	t.Run("trash: delete, restore, delete for good", func(t *testing.T) {
		s.t = t
		s.upload("trash-me.txt", "bye")
		if err := client.DeleteResource(ctx, s.p("trash-me.txt"), false); err != nil {
			t.Fatalf("delete to trash: %v", err) // the 204 bug of v1.3
		}
		item := s.trashItem("trash-me.txt")
		meta, err := client.GetTrashResourceMetadata(ctx, item, nil)
		if err != nil || meta.OriginPath != s.p("trash-me.txt") {
			t.Errorf("trash metadata: %+v %v", meta, err)
		}
		link, err := client.RestoreFromTrashWithOptions(ctx, item, &RestoreOptions{Overwrite: true})
		if err != nil {
			t.Fatal(err)
		}
		s.waitOperation(link)
		if !s.exists("trash-me.txt") {
			t.Fatal("not restored")
		}

		if err := client.DeleteResource(ctx, s.p("trash-me.txt"), false); err != nil {
			t.Fatal(err)
		}
		item = s.trashItem("trash-me.txt")
		if err := client.EmptyTrash(ctx, item, false); err != nil {
			t.Fatalf("delete one trash item: %v", err)
		}

		// A permanent delete must skip the trash (the "permanent" bug of v1.3).
		s.upload("gone.txt", "gone")
		if err := client.DeleteResourceWithOptions(ctx, s.p("gone.txt"), &DeleteOptions{Permanently: true}); err != nil {
			t.Fatal(err)
		}
		list, err := client.ListTrashResourcesWithOptions(ctx, "", &TrashListOptions{Limit: 50, Sort: "-deleted"})
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range list.Items {
			if it.OriginPath == s.p("gone.txt") {
				t.Error("a permanent delete ended up in the trash")
			}
		}
	})

	t.Run("the guard holds", func(t *testing.T) {
		s.t = t
		_, errResp := client.CreateDir(ctx, "disk:/should-never-exist-"+time.Now().Format("150405"))
		if errResp == nil || !strings.Contains(errResp.Error, "guard") {
			t.Errorf("guard let a change outside the sandbox through: %+v", errResp)
		}
		if err := client.EmptyTrash(ctx, "", false); err == nil || !strings.Contains(err.Error(), "guard") {
			t.Errorf("guard let the whole trash be emptied: %v", err)
		}
	})
}
