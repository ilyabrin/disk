# Yandex.Disk API Client - golang library

[![Run Tests](https://github.com/ilyabrin/disk/actions/workflows/test.yml/badge.svg)](https://github.com/ilyabrin/disk/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/ilyabrin/disk/badge.svg?branch=release)](https://coveralls.io/github/ilyabrin/disk?branch=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilyabrin/disk)](https://goreportcard.com/report/github.com/ilyabrin/disk)
[![Go Reference](https://pkg.go.dev/badge/github.com/ilyabrin/disk.svg)](https://pkg.go.dev/github.com/ilyabrin/disk)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/ilyabrin/disk)

A comprehensive Go client library for the [Yandex.Disk REST API](https://yandex.ru/dev/disk/rest/). This library provides a clean, idiomatic Go interface for interacting with Yandex.Disk cloud storage service.

## Features

- ✅ Complete API coverage for Yandex.Disk REST API
- ✅ Context-aware operations with timeout support
- ✅ Configurable HTTP client with retry mechanisms
- ✅ Comprehensive error handling
- ✅ Built-in logging with configurable levels
- ✅ Batch operations support
- ✅ File upload/download with progress tracking
- ✅ Public link management
- ✅ Trash operations
- ✅ Resource metadata operations
- ✅ Pagination support for large datasets

## Installation

```bash
go get -u github.com/ilyabrin/disk
```

## Quick Start

### Authentication

First, you need to obtain an access token from Yandex. You can get one by following the [Yandex OAuth documentation](https://yandex.ru/dev/disk/rest/reference/access-token.html).

Set your access token as an environment variable:

```bash
export YANDEX_DISK_ACCESS_TOKEN=your_access_token_here
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ilyabrin/disk"
)

func main() {
    // Create a new client (reads token from environment variable)
    client := disk.New()
    
    // Or create with explicit token
    // client := disk.New("your_access_token_here")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Get disk information
    diskInfo, err := client.DiskInfo(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total space: %d bytes\n", diskInfo.TotalSpace)
    fmt.Printf("Used space: %d bytes\n", diskInfo.UsedSpace)
    fmt.Printf("Available space: %d bytes\n", diskInfo.TotalSpace-diskInfo.UsedSpace)
}
```

## Advanced Usage

### Custom Client Configuration

```go
config := &disk.ClientConfig{
    DefaultTimeout:     60 * time.Second,
    MaxRetries:        5,
    EnableDebugLogging: true,
    Logger: &disk.LoggerConfig{
        Level: disk.LogLevelDebug,
        Output: os.Stdout,
    },
}

client := disk.NewWithConfig("your_token", config)
```

### File Operations

```go
// Upload a file
uploadURL, err := client.GetUploadURL(ctx, "/remote/path/file.txt", false)
if err != nil {
    log.Fatal(err)
}

file, err := os.Open("local/file.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = client.UploadFile(ctx, uploadURL.Href, file)
if err != nil {
    log.Fatal(err)
}

// Download a file
downloadURL, err := client.GetDownloadURL(ctx, "/remote/path/file.txt")
if err != nil {
    log.Fatal(err)
}

resp, err := http.Get(downloadURL.Href)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Save to local file
localFile, err := os.Create("downloaded_file.txt")
if err != nil {
    log.Fatal(err)
}
defer localFile.Close()

_, err = io.Copy(localFile, resp.Body)
if err != nil {
    log.Fatal(err)
}
```

### Resource Management

```go
// Create a directory
err := client.CreateFolder(ctx, "/new/directory/path")
if err != nil {
    log.Fatal(err)
}

// List directory contents
resources, err := client.GetResourcesList(ctx, "/path/to/directory", &disk.ResourcesListParams{
    Limit: 100,
    Sort:  "name",
})
if err != nil {
    log.Fatal(err)
}

for _, resource := range resources.Items {
    fmt.Printf("%s: %s (%d bytes)\n", resource.Type, resource.Name, resource.Size)
}

// Move a file
err = client.MoveResource(ctx, "/old/path/file.txt", "/new/path/file.txt", false)
if err != nil {
    log.Fatal(err)
}

// Copy a file
err = client.CopyResource(ctx, "/source/file.txt", "/destination/file.txt", false)
if err != nil {
    log.Fatal(err)
}
```

### Public Links

```go
// Publish a resource
publicResource, err := client.PublishResource(ctx, "/path/to/file.txt")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Public URL: %s\n", publicResource.PublicURL)

// Get public resources list
publicResources, err := client.GetPublicResourcesList(ctx, &disk.PublicResourcesParams{
    Limit: 50,
    Type:  "file",
})
if err != nil {
    log.Fatal(err)
}
```

### Trash Operations

```go
// Delete a resource (move to trash)
err := client.DeleteResource(ctx, "/path/to/file.txt", false)
if err != nil {
    log.Fatal(err)
}

// List trash contents
trashResources, err := client.GetTrashResourcesList(ctx, &disk.TrashResourcesParams{
    Limit: 100,
})
if err != nil {
    log.Fatal(err)
}

// Restore from trash
err = client.RestoreTrashResource(ctx, "/trash/path/file.txt", "/restored/path/file.txt", false)
if err != nil {
    log.Fatal(err)
}

// Permanently delete from trash
err = client.DeleteTrashResource(ctx, "/trash/path/file.txt")
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The library provides detailed error information:

```go
diskInfo, err := client.DiskInfo(ctx)
if err != nil {
    var diskErr *disk.Error
    if errors.As(err, &diskErr) {
        fmt.Printf("API Error: %s (Code: %s)\n", diskErr.Message, diskErr.ErrorCode)
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Testing

Run the test suite:

```bash
go test -v ./...
```

Run tests with coverage:

```bash
go test -v -cover ./...
```

## API Reference

For detailed API documentation, visit [pkg.go.dev](https://pkg.go.dev/github.com/ilyabrin/disk).

### Core Methods

- `DiskInfo(ctx)` - Get disk space information
- `GetResourcesList(ctx, path, params)` - List directory contents
- `CreateFolder(ctx, path)` - Create a directory
- `MoveResource(ctx, from, to, overwrite)` - Move/rename a resource
- `CopyResource(ctx, from, to, overwrite)` - Copy a resource
- `DeleteResource(ctx, path, permanently)` - Delete a resource
- `GetUploadURL(ctx, path, overwrite)` - Get upload URL
- `GetDownloadURL(ctx, path)` - Get download URL
- `PublishResource(ctx, path)` - Make resource public
- `UnpublishResource(ctx, path)` - Make resource private

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Yandex.Disk REST API Documentation](https://yandex.ru/dev/disk/rest/)
- The Go community for excellent HTTP client patterns

---

**Note**: This library is not officially affiliated with Yandex. Yandex.Disk is a trademark of Yandex LLC.
