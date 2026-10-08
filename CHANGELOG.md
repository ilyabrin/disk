# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project uses [semantic versioning](https://semver.org/).

## [Unreleased]

## [v1.5.0] - 2026-10-08

### Added

- Public links protected with a password or an expiry: `PublishResourceWithSettings`, `UpdatePublicSettings` and `GetPublicSettings`, with `PublicSettings`. Built on what the live API does, which differs from its reference; forbidding downloads is not available through the public API
- An integration suite against the real API, behind the `integration` build tag. It works in a folder it deletes, behind a guard that refuses any change outside it

### Fixed

- **Request bodies went out chunked, and the API read them as empty.** `UpdateMetadata` never stored anything
- Batch copy and move ignored `Overwrite`
- Waiting for a batch failed with "Operation not found" when a copy or move had finished at once

## [v1.4.0] - 2026-10-08

### Added

- `*WithOptions` variants for copy, move, delete, restore, the trash list, last uploaded and public resources, and upload by URL: `overwrite`, `force_async`, `md5`, filters, sorting and previews
- `SavePublicResourceOptions.SavePath` and `ForceAsync`
- `GetLinkForUploadWithOverwrite`
- Fields the API returns: paid file size limit, mail, disk and photo sizes, overdraft warning, monthly traffic limit and more in `Disk`; `Resource.Sizes`; `User.IsChild` and `RegTime`
- Back from v1.0.1: the `Scans`, `Attach`, `Messenger` and `Calendar` system folders, and GPS coordinates in `Exif`, as a `Coordinate` type that tolerates odd values

### Fixed

- **`DeleteResource` reported every successful delete as an error.** The API answers 204, or 202 for a large folder, and only 200 was accepted. Batch deletes went through the same path
- **A permanent delete went to the trash:** the parameter was sent as `permanent` instead of `permanently`
- **Uploading over an existing file failed with 409:** overwrite was sent as an `X-Overwrite` header the API does not know, instead of `overwrite=true` on the link request
- **`OperationStatus` asked the wrong URL:** the operation ID belongs in the path, `operations/{id}`

### Changed

- Every exported type, method and constant is documented, and the package has an overview on pkg.go.dev, with examples. The linter now requires a doc comment on new exported API

## [v1.3.0] - 2026-10-07

### Fixed

- **Sizes and revisions overflowed `int` on 32-bit systems** (386, 32-bit ARM). A Disk or a file over 2 GB, and every revision, failed to decode, so even `DiskInfo` returned an error there

### Changed

- **Breaking:** `Resource.Size`, `Resource.Revision`, and `Disk.TotalSpace`, `UsedSpace`, `TrashSize`, `MaxFileSize` and `Revision` are `int64` instead of `int`. Code that keeps one of them in an `int` needs a conversion; the README shows how
- CI also runs the tests as 386

## [v1.2.3] - 2026-10-07

### Fixed

- **`ListTrashResources` always returned an empty list against the real API.** Yandex returns the trash contents under `_embedded`, like an ordinary folder; they are now read from there, and a list at the top level is still accepted

### Added

- `TrashResourceList.Total`, the number of items in the trash

### Changed

- CI checks `gofmt`, `go mod tidy` and a pinned golangci-lint; gosec skips `examples/`
- `CONTRIBUTING.md`, `SECURITY.md`, a code of conduct, and issue and pull request templates
- The default branch is now `main`

## [v1.2.2] - 2026-10-06

### Fixed

- **Debug logging of request headers was inverted.** With `LoggerConfig.Verbose` on, ordinary headers were printed as `[sanitized]` while `Authorization` was printed masked, and `SanitizeAuth: false` hid every header. Headers are now logged in a stable order
- Masking no longer keeps any part of a credential. `Authorization` keeps only its scheme, such as `OAuth`

### Changed

- `.golangci.yml` is a working golangci-lint v2 config
- The README documents `CreateDirAll`, `ErrorResponse.StatusCode` and `AlreadyExists()`

## [v1.2.1] - 2026-09-19

### Fixed

- **Uploads and downloads no longer die at `DefaultTimeout`.** It was applied as an absolute deadline to the whole transfer, so with the 30 second default anything slower than that was aborted. Transfers are now bounded by the request context, or 30 minutes when it has no deadline

### Added

- `CreateDirAll` creates a nested path in one call, like `os.MkdirAll`
- `ErrorResponse.StatusCode`, the HTTP status behind an error
- `ErrorResponse.AlreadyExists()` tells an "already there" conflict apart from other errors

## [v1.2.0] - 2026-08-10

### Added

- `fields`, `media_type`, `sort`, `preview_size` and `preview_crop` parameters: `ResourceOptions.Fields`, `FilesOptions` and `GetSortedFilesWithOptions`
- Options for public resources: `GetMetadataForPublicResourceWithOptions`, `GetDownloadURLForPublicResourceAt` and `SavePublicResourceWithOptions`
- `GetOperationStatus` and `OperationIDFromHref` for following background operations
- `ClientConfig.BaseURL` to point the client at another endpoint

### Fixed

- `EmptyTrash` sent `force_async=false` when asked for asynchronous deletion
- `RestoreFromTrash` dropped the link returned with 201
- Path validation rejected names containing `..`, such as `report..final.pdf`, and applied Windows rules to Disk paths
- Multipart upload changed the shared `HTTPClient.Timeout`, a data race for concurrent callers
- Progress callbacks divided by zero on empty files; MIME detection could truncate on a short read

### Changed

- **The minimum Go version is 1.23** (was 1.20)
- `MaxRetries` works: requests without a body retry on connection errors, 429 and 5xx, with exponential backoff
- `EnableDebugLogging` is honoured
- `WaitForBatchOperation` polls the operations instead of only logging
- Tests run with `-race`; both READMEs were rewritten

## [v1.1.1] - 2026-05-08

### Added

- `ResourceOptions` for `GetMetadataWithOptions`: page size, offset, sort order, preview size and fields of a folder listing

## [v1.1.0] - 2026-01-13

### Added

- Batch operations: delete, copy, move and update metadata of many resources in parallel, with status tracking and retries
- Pagination: `PaginationOptions`, `Paged` results and iterators over all pages
- Uploads from a local path, with progress reporting and chunked upload for large files
- Trash support reworked: list, restore, empty, and metadata of deleted items
- Structured logging with levels, in place of `log.Fatal` calls
- Request timeouts through `context.Context`
- Path validation
- A Russian README and runnable examples

### Removed

- The fields added in v1.0.1 (`Disk.PaidMaxFileSize`, the `Attach`, `Messenger`, `Calendar` and `Scans` system folders, and the GPS coordinates in `Exif`) were lost when the history was rewritten for this release

## [v1.0.1] - 2022-11-11

### Added

- `Disk.PaidMaxFileSize`
- `Attach`, `Messenger`, `Calendar` and `Scans` in `SystemFolders`
- `Longitude` and `Latitude` in `Exif`

## [v.1.0.0] - 2022-11-06

### Added

- Trash and background operation support, tests with recorded API answers, CI and a Dockerfile

## [v.0.1-alpha] - 2020-11-09

### Added

- First release: disk info, metadata, creating, copying, moving and deleting resources, upload and download links, publishing, public resources and the list of recently uploaded files

[Unreleased]: https://github.com/ilyabrin/disk/compare/v1.5.0...HEAD
[v1.5.0]: https://github.com/ilyabrin/disk/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/ilyabrin/disk/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/ilyabrin/disk/compare/v1.2.3...v1.3.0
[v1.2.3]: https://github.com/ilyabrin/disk/compare/v1.2.2...v1.2.3
[v1.2.2]: https://github.com/ilyabrin/disk/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/ilyabrin/disk/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/ilyabrin/disk/compare/v1.1.1...v1.2.0
[v1.1.1]: https://github.com/ilyabrin/disk/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/ilyabrin/disk/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/ilyabrin/disk/compare/v.1.0.0...v1.0.1
[v.1.0.0]: https://github.com/ilyabrin/disk/compare/v.0.1-alpha...v.1.0.0
[v.0.1-alpha]: https://github.com/ilyabrin/disk/releases/tag/v.0.1-alpha
