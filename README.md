<div align="center">

# disk

**An idiomatic Go client for the [Yandex Disk REST API](https://yandex.ru/dev/disk-api/doc/en/).**

[![Go Reference](https://pkg.go.dev/badge/github.com/ilyabrin/disk.svg)](https://pkg.go.dev/github.com/ilyabrin/disk)
[![Run Tests](https://github.com/ilyabrin/disk/actions/workflows/test.yml/badge.svg)](https://github.com/ilyabrin/disk/actions/workflows/test.yml)
[![Security Checks](https://github.com/ilyabrin/disk/actions/workflows/security.yml/badge.svg)](https://github.com/ilyabrin/disk/actions/workflows/security.yml)
[![Coverage Status](https://coveralls.io/repos/github/ilyabrin/disk/badge.svg?branch=release)](https://coveralls.io/github/ilyabrin/disk?branch=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilyabrin/disk)](https://goreportcard.com/report/github.com/ilyabrin/disk)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ilyabrin/disk)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

[Русская версия](./README.ru.md) · [API reference](https://pkg.go.dev/github.com/ilyabrin/disk) · [Pagination guide](./PAGINATION.md)

</div>

---

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [Authentication](#authentication)
- [Disk info](#disk-info)
- [Files and folders](#files-and-folders)
- [Upload](#upload)
- [Download](#download)
- [Public resources](#public-resources)
- [Trash](#trash)
- [Pagination](#pagination)
- [Batch operations](#batch-operations)
- [Configuration](#configuration)
- [Logging](#logging)
- [Error handling](#error-handling)
- [Async operations](#async-operations)
- [Examples](#examples)
- [Development](#development)
- [License](#license)

---

## Install

```bash
go get github.com/ilyabrin/disk
```

> [!NOTE]
> Requires Go 1.23 or newer. The library has no runtime dependencies beyond the standard library.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ilyabrin/disk"
)

func main() {
	// Reads YANDEX_DISK_ACCESS_TOKEN when no token is passed explicitly.
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	info, err := client.DiskInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Used %s of %s\n",
		disk.FormatFileSize(int64(info.UsedSpace)),
		disk.FormatFileSize(int64(info.TotalSpace)),
	)
}
```

## Authentication

The client authenticates with an OAuth token. Get one from the
[Yandex OAuth console](https://oauth.yandex.ru/) with the `cloud_api:disk.*` scopes.

```go
client, err := disk.New("y0_AgAAAA...")       // explicit token
client, err := disk.New()                     // from $YANDEX_DISK_ACCESS_TOKEN
```

> [!WARNING]
> Never commit tokens. The built-in logger redacts `Authorization` headers, but
> your own logging is your responsibility.

## Disk info

```go
info, err := client.DiskInfo(ctx)
if err != nil {
	log.Fatal(err)
}

fmt.Println("Total:", info.TotalSpace)
fmt.Println("Used:", info.UsedSpace)
fmt.Println("Trash:", info.TrashSize)
fmt.Println("Downloads folder:", info.SystemFolders.Downloads)
```

## Files and folders

<details open>
<summary><b>Metadata</b></summary>

```go
resource, errResp := client.GetMetadata(ctx, "/Documents/report.pdf")
if errResp != nil {
	log.Fatal(errResp.Error)
}
fmt.Println(resource.Name, resource.Size, resource.MimeType)
```

Listing a folder returns its contents in `_embedded`. Use `GetMetadataWithOptions`
to paginate, sort, and trim the response:

```go
folder, errResp := client.GetMetadataWithOptions(ctx, "/Photos", &disk.ResourceOptions{
	Limit:       100,
	Offset:      0,
	Sort:        "-modified",                       // "-" reverses the order
	PreviewSize: "M",
	PreviewCrop: true,
	Fields:      []string{"name", "_embedded.items.name", "_embedded.items.size"},
})

for _, item := range folder.Embedded.Items {
	fmt.Println(item.Type, item.Name)
}
```

</details>

<details>
<summary><b>Create, copy, move, delete</b></summary>

```go
// Create a folder
link, errResp := client.CreateDir(ctx, "/Reports")

// Copy
link, errResp = client.CopyResource(ctx, "/a.txt", "/backup/a.txt")

// Move or rename
link, errResp = client.MoveResource(ctx, "/a.txt", "/archive/a-2026.txt")

// Delete (to trash)
err := client.DeleteResource(ctx, "/a.txt", false)

// Delete permanently
err = client.DeleteResource(ctx, "/a.txt", true)
```

</details>

<details>
<summary><b>Custom properties</b></summary>

```go
resource, errResp := client.UpdateMetadata(ctx, "/report.pdf",
	map[string]map[string]string{
		"custom_properties": {
			"project": "apollo",
			"status":  "final",
		},
	})
```

</details>

<details>
<summary><b>Flat file list and recent uploads</b></summary>

```go
// All files, newest first, images and video only
files, errResp := client.GetSortedFilesWithOptions(ctx,
	&disk.PaginationOptions{Limit: 50},
	&disk.FilesOptions{
		MediaType: []string{"image", "video"},
		Sort:      "-created",
	})

// Last uploaded resources
recent, errResp := client.GetLastUploadedResources(ctx)
```

</details>

## Upload

| Method | Use for |
| --- | --- |
| `UploadFileFromPath` | Any local file, with full option control |
| `UploadFileFromPathWithProgress` | Small/medium files with a progress bar |
| `UploadLargeFileFromPath` | Large files, chunked progress reporting |
| `UploadFile` | Server-side fetch: Yandex downloads a URL for you |

```go
resource, err := client.UploadFileFromPath(ctx, "./report.pdf", "/Documents/report.pdf",
	&disk.UploadOptions{Overwrite: true})
```

With progress:

```go
resource, err := client.UploadFileFromPathWithProgress(ctx,
	"./video.mp4", "/Videos/video.mp4", true,
	func(p disk.UploadProgress) {
		fmt.Printf("\r%.1f%% (%s / %s)",
			p.Percentage,
			disk.FormatFileSize(p.BytesUploaded),
			disk.FormatFileSize(p.TotalBytes),
		)
	})
```

Large files, reporting once per 10 MB chunk:

```go
resource, err := client.UploadLargeFileFromPath(ctx,
	"./archive.zip", "/Backups/archive.zip", 10,
	func(p disk.UploadProgress) {
		log.Printf("uploaded %s", disk.FormatFileSize(p.BytesUploaded))
	})
```

Let Yandex fetch a remote URL directly, without routing bytes through your process:

```go
link, errResp := client.UploadFile(ctx, "/Downloads/image.jpg", "https://example.com/image.jpg")
```

> [!TIP]
> `UploadFile` is asynchronous — poll the returned link with
> [`GetOperationStatus`](#async-operations) to find out when the file has landed.

## Download

```go
err := client.DownloadFileToPath(ctx, "/Photos/image.jpg", "./image.jpg",
	&disk.DownloadOptions{Overwrite: true})
```

With progress:

```go
err := client.DownloadFileToPathWithProgress(ctx,
	"/Videos/video.mp4", "./video.mp4", true,
	func(p disk.DownloadProgress) {
		if p.TotalBytes > 0 {
			fmt.Printf("\r%.1f%%", p.Percentage)
		}
	})
```

Need the raw link instead (for a CDN, a browser redirect, or your own transfer code)?

```go
link, errResp := client.GetDownloadURL(ctx, "/Photos/image.jpg")
fmt.Println(link.Href) // short-lived, single-use
```

## Public resources

```go
// Publish and unpublish
link, errResp := client.PublishResource(ctx, "/Photos/image.jpg")
link, errResp = client.UnpublishResource(ctx, "/Photos/image.jpg")

// Everything you have published
list, errResp := client.GetPublicResources(ctx)
```

Reading someone else's published resource by key or URL:

```go
resource, errResp := client.GetMetadataForPublicResource(ctx, "https://yadi.sk/d/abc123")

// Browse inside a published folder
resource, errResp = client.GetMetadataForPublicResourceWithOptions(ctx, "https://yadi.sk/d/abc123",
	&disk.PublicResourceOptions{
		Path:  "/subfolder",
		Sort:  "name",
		Limit: 50,
	})

// Download a specific file from a published folder
link, errResp := client.GetDownloadURLForPublicResourceAt(ctx, "https://yadi.sk/d/abc123", "/subfolder/file.txt")

// Save it into your own Downloads folder under a new name
link, errResp = client.SavePublicResourceWithOptions(ctx, "https://yadi.sk/d/abc123",
	&disk.SavePublicResourceOptions{Path: "/subfolder/file.txt", Name: "copy.txt"})
```

## Trash

```go
// Browse
trash, err := client.ListTrashResources(ctx, "", 100, 0)

// Metadata for one item
item, err := client.GetTrashResourceMetadata(ctx, "report.pdf", nil)

// Restore, optionally renaming
link, err := client.RestoreFromTrash(ctx, "report.pdf", false, "report-restored.pdf")

// Empty a single path, or the whole trash with ""
err = client.EmptyTrash(ctx, "", false)
```

> [!NOTE]
> Pass `forceAsync: true` to `EmptyTrash` to make the API always answer `202` with
> an operation link instead of blocking on a large deletion.

## Pagination

Three styles are available, from lowest to highest level. See [PAGINATION.md](./PAGINATION.md) for the full guide.

<details open>
<summary><b>1. Explicit limit/offset</b></summary>

```go
files, errResp := client.GetSortedFilesWithPagination(ctx, &disk.PaginationOptions{
	Limit:  50,
	Offset: 100,
})
```

</details>

<details>
<summary><b>2. Paged results with metadata</b></summary>

```go
page, errResp := client.GetSortedFilesPaged(ctx, &disk.PaginationOptions{Limit: 50})
fmt.Println(page.Pagination.HasMore, page.Pagination.NextOffset)
```

</details>

<details>
<summary><b>3. Iterator</b></summary>

```go
it := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 100})

for it.HasNext() {
	page, err := it.Next(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range page.Items {
		fmt.Println(file.Name)
	}
}
```

</details>

## Batch operations

Batch helpers run operations concurrently and collect per-item results instead of
failing on the first error.

```go
status, err := client.BatchDeleteFiles(ctx,
	[]string{"/tmp/a.txt", "/tmp/b.txt", "/tmp/c.txt"},
	&disk.BatchDeleteOptions{
		BatchOptions: disk.BatchOptions{
			MaxConcurrency:  4,
			ContinueOnError: true,
			Progress: func(s disk.BatchOperationStatus) {
				fmt.Printf("\r%d/%d", s.Completed, s.Total)
			},
		},
		Permanently: false,
	})

fmt.Println(status.GetSummary())

for _, failure := range status.GetFailedOperations() {
	log.Printf("%s: %v", failure.Path, failure.Error)
}

// Retry only what failed
status, err = client.RetryFailedOperations(ctx, status, 2)
```

Available: `BatchDeleteFiles`, `BatchCopyFiles`, `BatchMoveFiles`,
`BatchUpdateMetadata`, plus the convenience wrappers `BatchRenameFiles`,
`BatchMoveToDirectory`, `BatchCopyToDirectory` and the `*Simple` variants.

## Configuration

```go
client, err := disk.NewWithConfig(&disk.ClientConfig{
	DefaultTimeout:     60 * time.Second,
	MaxRetries:         3,
	RetryBackoff:       200 * time.Millisecond,
	EnableDebugLogging: true,
	Logger:             disk.DefaultLoggerConfig(),
}, "your-token")
```

| Field | Default | Meaning |
| --- | --- | --- |
| `DefaultTimeout` | `30s` | Per-request timeout when the context carries no deadline |
| `MaxRetries` | `3` | Extra attempts for retryable requests |
| `RetryBackoff` | `200ms` | Base delay between retries; doubles each attempt |
| `EnableDebugLogging` | `false` | Switches the logger to `DEBUG` and verbose mode |
| `Logger` | see below | Logger configuration |
| `BaseURL` | Yandex API | Override the endpoint (tests, proxies) |

> [!IMPORTANT]
> Only requests **without a body** are retried — `GET`, `DELETE`, and the `PUT`/`POST`
> calls that carry their parameters in the query string. A request whose body is an
> `io.Reader` cannot be rewound, so it is sent exactly once. Retries fire on
> connection errors, `429`, and `5xx`.

Timeouts can also be set per call through the context:

```go
ctx, cancel := disk.WithTimeout(10 * time.Second)
defer cancel()

info, err := client.DiskInfo(ctx)
```

## Logging

```go
client.SetLogLevel(disk.DEBUG)   // DEBUG, INFO, WARN, ERROR, SILENT
client.SetVerbose(true)          // include request/response details
client.SetLogOutput(os.Stderr)   // any io.Writer
```

Sensitive header values (`Authorization`, tokens) are redacted before they reach
the log output.

## Error handling

The library uses two error conventions, and which one you get depends on the call:

| Return type | Where | How to handle |
| --- | --- | --- |
| `*ErrorResponse` | Resource, public and pagination calls | Non-`nil` means failure; read `.Error` and `.Description` |
| `error` | Disk info, upload, download, trash, batch | Standard Go handling, wrapped with `%w` |

```go
resource, errResp := client.GetMetadata(ctx, "/missing.txt")
if errResp != nil {
	log.Printf("%s: %s", errResp.Error, errResp.Description)
	return
}

if err := client.DownloadFileToPath(ctx, "/a.txt", "./a.txt", nil); err != nil {
	log.Fatal(err)
}
```

## Async operations

Copy, move, save-to-disk and trash operations may answer `202 Accepted` with a link
to a background operation. Poll it until it reports `success`:

```go
link, errResp := client.CopyResource(ctx, "/big-folder", "/backup/big-folder")
if errResp != nil {
	log.Fatal(errResp.Error)
}

for {
	// Accepts either an operation ID or the full href from the response.
	operation, err := client.GetOperationStatus(ctx, link.Href)
	if err != nil {
		log.Fatal(err)
	}
	if operation.Status != disk.OperationInProgress {
		fmt.Println("finished:", operation.Status)
		break
	}
	time.Sleep(time.Second)
}
```

For batch calls, `WaitForBatchOperation` does the polling for you:

```go
err := client.WaitForBatchOperation(ctx, status, time.Second)
```

## Examples

Runnable programs live in [examples/](./examples):

| Example | Shows |
| --- | --- |
| [demo](./examples/demo) | Disk info, metadata, folder and file operations |
| [upload](./examples/upload) | Uploads with progress reporting |
| [pagination](./examples/pagination) | All three pagination styles |

```bash
export YANDEX_DISK_ACCESS_TOKEN=your-token
go run ./examples/demo
```

## Development

```bash
go test ./...                                  # run the suite
go test -race -cover ./...                     # with the race detector
go vet ./...                                   # static checks
golangci-lint run                              # full lint (see .golangci.yml)
```

CI runs tests, `go vet`, CodeQL, `govulncheck` and gosec on every push and pull
request; gosec findings are published as code scanning alerts.

Contributions are welcome — open an issue or a pull request.

## License

[MIT](./LICENSE) © Ilya Brin
