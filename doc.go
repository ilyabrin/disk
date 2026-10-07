// Package disk is a client for the Yandex.Disk REST API: list, upload,
// download, copy, move, publish and delete files, and work with the trash.
//
// # Getting started
//
// The client signs requests with a user's OAuth token. Register an
// application at https://oauth.yandex.ru with the cloud_api:disk.* scopes
// and obtain a token for the user, then:
//
//	client, err := disk.New("y0_...") // or disk.New() to read YANDEX_DISK_ACCESS_TOKEN
//	if err != nil {
//		log.Fatal(err)
//	}
//	info, err := client.DiskInfo(ctx)
//
// Paths may be written as "disk:/Photos/cat.jpg" or "/Photos/cat.jpg"; the
// two are the same. The trash uses "trash:/".
//
// # Errors
//
// Many methods return an [*ErrorResponse] in place of an error: nil on
// success, otherwise its Error field carries the API's code, such as
// "DiskNotFoundError", and StatusCode the HTTP status. Others, such as
// [Client.DiskInfo] and the upload and download helpers, return a plain
// error.
//
// # Background operations
//
// Copying or moving a large folder can continue after the request
// returns. Such methods return a [Link] to the operation; poll it with
// [Client.GetOperationStatus] until it reports "success" or "failed".
//
// # Pages
//
// Lists are paged. Methods ending in WithPagination take a
// [PaginationOptions], those ending in Paged report whether more pages
// follow, and those ending in Iterator walk all pages for you.
//
// # Configuration
//
// [NewWithConfig] sets timeouts, retries, logging and the API endpoint; see
// [ClientConfig]. Requests without a body are retried on connection errors
// and on 429 and 5xx answers.
package disk
