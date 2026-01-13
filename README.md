# Yandex Disk Go Client Library

[🇷🇺 Русская версия](./README.ru.md)

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

A comprehensive, production-ready Go client library for the [Yandex Disk REST API](https://yandex.ru/dev/disk/rest/). This library provides a clean, idiomatic Go interface to interact with Yandex Disk cloud storage service.

## 🌟 Features

### Core Functionality

- ✅ **Complete API Coverage** - Full support for all Yandex Disk REST API endpoints
- 🔐 **OAuth2 Authentication** - Simple token-based authentication
- 📁 **File Operations** - Upload, download, copy, move, and delete files
- 📊 **Metadata Management** - Get and update file/folder metadata
- 🗑️ **Trash Management** - Move to trash, restore, and permanently delete
- 🌐 **Public Links** - Create and manage public links to files and folders
- 📦 **Disk Information** - Get disk space, quota, and system folders info

### Advanced Features

- 🔄 **Pagination Support** - Multiple pagination strategies (offset-based and iterator pattern)
- 📦 **Batch Operations** - Process multiple files efficiently with concurrent execution
- 📤 **Smart Upload** - Automatic handling of large files with progress tracking
- ⚡ **Context Support** - Full context.Context integration for timeouts and cancellation
- 📝 **Comprehensive Logging** - Built-in structured logging with multiple severity levels
- 🔒 **Security** - Path validation and sanitization to prevent attacks
- ⚙️ **Configurable** - Extensive configuration options for timeouts, retries, and more
- 🧪 **Well-Tested** - High test coverage with comprehensive unit tests

## 📋 Table of Contents

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Authentication](#-authentication)
- [Basic Usage](#-basic-usage)
  - [Disk Information](#disk-information)
  - [File Operations](#file-operations)
  - [Folder Operations](#folder-operations)
  - [Upload Files](#upload-files)
  - [Download Files](#download-files)
  - [Public Resources](#public-resources)
  - [Trash Management](#trash-management)
- [Advanced Features](#-advanced-features)
  - [Pagination](#pagination)
  - [Batch Operations](#batch-operations)
  - [Progress Tracking](#progress-tracking)
  - [Custom Configuration](#custom-configuration)
  - [Logging](#logging)
- [Error Handling](#️-error-handling)
- [Examples](#-examples)
- [API Reference](#-api-reference)
- [Testing](#-testing)
- [Contributing](#-contributing)
- [License](#-license)

## 📦 Installation

```bash
go get github.com/ilyabrin/disk
```

**Requirements:**

- Go 1.20 or higher

## 🚀 Quick Start

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
    // Create a new client with your OAuth token
    client := disk.NewClient("YOUR_OAUTH_TOKEN")

    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Get disk information
    diskInfo, err := client.GetDisk(ctx)
    if err != nil {
        log.Fatalf("Failed to get disk info: %v", err)
    }

    fmt.Printf("Total Space: %d bytes\n", diskInfo.TotalSpace)
    fmt.Printf("Used Space: %d bytes\n", diskInfo.UsedSpace)
    fmt.Printf("Available: %d bytes\n", diskInfo.TotalSpace-diskInfo.UsedSpace)
}
```

## 🔐 Authentication

To use this library, you need an OAuth token from Yandex. Here's how to get one:

1. Go to [Yandex OAuth](https://oauth.yandex.ru/)
2. Register your application
3. Request access to `cloud_api:disk.read` and `cloud_api:disk.write` scopes
4. Obtain your OAuth token

```go
// Create a client with your token
client := disk.NewClient("YOUR_OAUTH_TOKEN")

// Or with custom configuration
config := disk.DefaultClientConfig()
config.DefaultTimeout = 60 * time.Second
client = disk.NewClientWithConfig("YOUR_OAUTH_TOKEN", config)
```

## 💡 Basic Usage

### Disk Information

```go
// Get disk information
diskInfo, err := client.GetDisk(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total Space: %d\n", diskInfo.TotalSpace)
fmt.Printf("Used Space: %d\n", diskInfo.UsedSpace)
fmt.Printf("Trash Size: %d\n", diskInfo.TrashSize)
fmt.Printf("Is Paid: %t\n", diskInfo.IsPaid)
```

### File Operations

#### Get File Metadata

```go
// Get metadata for a file or folder
resource, errResp := client.GetMetadata(ctx, "/path/to/file.txt")
if errResp != nil {
    log.Fatalf("Error: %s", errResp.Error)
}

fmt.Printf("Name: %s\n", resource.Name)
fmt.Printf("Size: %d bytes\n", resource.Size)
fmt.Printf("Type: %s\n", resource.Type)
fmt.Printf("Modified: %s\n", resource.Modified)
```

#### Copy File

```go
// Copy a file or folder
link, err := client.CopyResource(ctx, "/source/file.txt", "/destination/file.txt", false)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Copy operation link: %s\n", link.Href)
```

#### Move/Rename File

```go
// Move or rename a file
link, err := client.MoveResource(ctx, "/old/path/file.txt", "/new/path/file.txt", false)
if err != nil {
    log.Fatal(err)
}
```

#### Delete File

```go
// Delete a file (move to trash)
err := client.DeleteResource(ctx, "/path/to/file.txt", false)
if err != nil {
    log.Fatal(err)
}

// Permanently delete a file
err = client.DeleteResource(ctx, "/path/to/file.txt", true)
```

### Folder Operations

#### Create Folder

```go
// Create a new folder
link, err := client.CreateFolder(ctx, "/path/to/new/folder")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Folder created: %s\n", link.Href)
```

#### List Folder Contents

```go
// Get folder contents
resource, errResp := client.GetMetadata(ctx, "/path/to/folder")
if errResp != nil {
    log.Fatal(errResp.Error)
}

// Iterate through items
if resource.Embedded != nil {
    for _, item := range resource.Embedded.Items {
        fmt.Printf("- %s (%s)\n", item.Name, item.Type)
    }
}
```

### Upload Files

#### Simple Upload

```go
// Upload a small file
options := &disk.UploadOptions{
    Overwrite: true,
    Progress: func(progress disk.UploadProgress) {
        fmt.Printf("Uploaded: %.2f%%\n", progress.Percentage)
    },
}

resource, err := client.UploadFileFromPath(ctx, "local/file.txt", "/disk/file.txt", options)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File uploaded: %s\n", resource.Name)
```

#### Upload Large File

```go
// Upload a large file with automatic chunking
resource, err := client.UploadLargeFileFromPath(ctx, "large-file.zip", "/disk/large-file.zip", nil)
if err != nil {
    log.Fatal(err)
}
```

#### Upload from Reader

```go
// Upload from io.Reader
file, _ := os.Open("file.txt")
defer file.Close()

resource, err := client.UploadFile(ctx, file, "/disk/file.txt", options)
```

### Download Files

```go
// Download a file
err := client.DownloadFile(ctx, "/disk/file.txt", "local/file.txt")
if err != nil {
    log.Fatal(err)
}

// Or get download URL
link, errResp := client.GetDownloadURL(ctx, "/disk/file.txt")
if errResp != nil {
    log.Fatal(errResp.Error)
}
fmt.Printf("Download URL: %s\n", link.Href)
```

### Public Resources

#### Publish a Resource

```go
// Make a file or folder public
link, err := client.PublishResource(ctx, "/path/to/file.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Public URL: %s\n", link.Href)
```

#### Unpublish a Resource

```go
// Remove public access
link, err := client.UnpublishResource(ctx, "/path/to/file.txt")
if err != nil {
    log.Fatal(err)
}
```

#### Access Public Resource

```go
// Get metadata for a public resource
resource, errResp := client.GetMetadataForPublicResource(ctx, "public-key")
if errResp != nil {
    log.Fatal(errResp.Error)
}

// Download public resource
link, errResp := client.GetDownloadURLForPublicResource(ctx, "public-key")
```

#### List Public Resources

```go
// Get all public resources
options := &disk.PaginationOptions{Limit: 20}
pagedResources, errResp := client.GetPublicResourcesPaged(ctx, options)
if errResp != nil {
    log.Fatal(errResp.Error)
}

for _, resource := range pagedResources.Items {
    fmt.Printf("Public: %s (%s)\n", resource.Name, resource.PublicURL)
}
```

### Trash Management

#### Move to Trash

```go
// Delete file (moves to trash)
err := client.DeleteResource(ctx, "/path/to/file.txt", false)
```

#### List Trash Contents

```go
// List items in trash
trashList, err := client.ListTrashResources(ctx, "", 20, 0)
if err != nil {
    log.Fatal(err)
}

for _, item := range trashList.Embedded.Items {
    fmt.Printf("Trash: %s (deleted: %s)\n", item.Name, item.Deleted)
}
```

#### Restore from Trash

```go
// Restore a file from trash
link, err := client.RestoreFromTrash(ctx, "/path/to/file.txt", false, "")
if err != nil {
    log.Fatal(err)
}
```

#### Empty Trash

```go
// Permanently delete all trash
link, err := client.EmptyTrash(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### Permanently Delete from Trash

```go
// Delete specific item from trash permanently
link, err := client.DeleteFromTrash(ctx, "/path/to/file.txt")
```

## 🔥 Advanced Features

### Pagination

The library provides comprehensive pagination support with multiple strategies. See [PAGINATION.md](./PAGINATION.md) for detailed documentation.

#### Basic Pagination

```go
// Get files with pagination
options := &disk.PaginationOptions{
    Limit:  20,
    Offset: 0,
}
files, errResp := client.GetSortedFilesWithPagination(ctx, options)
```

#### Enhanced Pagination with Metadata

```go
// Get pagination info
pagedFiles, errResp := client.GetSortedFilesPaged(ctx, options)
if errResp == nil {
    fmt.Printf("Total items: %d\n", len(pagedFiles.Items))
    fmt.Printf("Has more: %t\n", pagedFiles.Pagination.HasMore)
    if pagedFiles.Pagination.HasMore {
        fmt.Printf("Next offset: %d\n", pagedFiles.Pagination.NextOffset)
    }
}
```

#### Iterator Pattern

```go
// Use iterator for automatic pagination
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 50})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        log.Printf("Error: %v", err)
        break
    }
    
    for _, file := range page.FilesResourceList.Items {
        fmt.Printf("File: %s (%d bytes)\n", file.Name, file.Size)
    }
    
    // Rate limiting
    time.Sleep(200 * time.Millisecond)
}
```

### Batch Operations

Process multiple files efficiently with concurrent execution.

#### Batch Delete

```go
paths := []string{
    "/file1.txt",
    "/file2.txt",
    "/folder/file3.txt",
}

options := &disk.BatchDeleteOptions{
    BatchOptions: disk.BatchOptions{
        MaxConcurrency:  5,
        ContinueOnError: true,
        Progress: func(status disk.BatchOperationStatus) {
            fmt.Printf("Progress: %d/%d (%.1f%%)\n", 
                status.Completed, status.Total, status.Percentage)
        },
    },
    Permanently: false,
}

status, err := client.BatchDeleteFiles(ctx, paths, options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Successful: %d, Failed: %d\n", status.Successful, status.Failed)
```

#### Batch Copy

```go
sourceDestMap := map[string]string{
    "/source/file1.txt": "/backup/file1.txt",
    "/source/file2.txt": "/backup/file2.txt",
}

options := &disk.BatchCopyMoveOptions{
    BatchOptions: disk.BatchOptions{
        MaxConcurrency: 3,
    },
    Overwrite: false,
}

status, err := client.BatchCopyFiles(ctx, sourceDestMap, options)
```

#### Batch Move

```go
status, err := client.BatchMoveFiles(ctx, sourceDestMap, options)
```

### Progress Tracking

Track upload/download progress in real-time.

```go
options := &disk.UploadOptions{
    Progress: func(progress disk.UploadProgress) {
        percentage := progress.Percentage
        bytes := progress.BytesUploaded
        total := progress.TotalBytes
        
        fmt.Printf("\rUploading: %.2f%% (%d/%d bytes)", 
            percentage, bytes, total)
    },
}

resource, err := client.UploadFileFromPath(ctx, localPath, remotePath, options)
```

### Custom Configuration

```go
// Create custom configuration
config := &disk.ClientConfig{
    DefaultTimeout:     60 * time.Second,
    MaxRetries:         3,
    EnableDebugLogging: true,
    Logger: &disk.LoggerConfig{
        Enabled:      true,
        Level:        disk.LogLevelInfo,
        IncludeTime:  true,
        ColorEnabled: true,
    },
}

client := disk.NewClientWithConfig("YOUR_TOKEN", config)
```

### Logging

The library includes comprehensive logging capabilities.

```go
// Access the logger
client.Logger.Info("Operation started")
client.Logger.Debug("Debug information: %v", data)
client.Logger.Error("An error occurred: %v", err)

// Change log level
client.Logger.SetLevel(disk.LogLevelDebug)

// Enable/disable logging
client.Logger.SetEnabled(true)

// Enable colored output
client.Logger.SetColorEnabled(true)
```

**Available Log Levels:**

- `LogLevelDebug` - Detailed debugging information
- `LogLevelInfo` - General informational messages
- `LogLevelWarn` - Warning messages
- `LogLevelError` - Error messages

## ⚠️ Error Handling

The library provides detailed error information through structured error responses.

```go
resource, errResp := client.GetMetadata(ctx, "/path/to/file")
if errResp != nil {
    fmt.Printf("Error: %s\n", errResp.Error)
    fmt.Printf("Description: %s\n", errResp.Description)
    fmt.Printf("Message: %s\n", errResp.Message)
    
    // Handle specific errors
    switch errResp.Error {
    case "DiskNotFoundError":
        fmt.Println("File or folder not found")
    case "UnauthorizedError":
        fmt.Println("Invalid or expired token")
    case "DiskPathPointsToExistentDirectoryError":
        fmt.Println("Path already exists")
    default:
        fmt.Printf("Unknown error: %s\n", errResp.Error)
    }
    return
}
```

**Common Error Types:**

- `UnauthorizedError` - Invalid or expired OAuth token
- `DiskNotFoundError` - Resource not found
- `DiskPathPointsToExistentDirectoryError` - Path already exists
- `FieldValidationError` - Invalid input parameters
- `LockedError` - Resource is locked
- `LimitExceededError` - Rate limit exceeded

## 📚 Examples

Complete working examples are available in the `examples/` directory:

- **[demo/main.go](./examples/demo/main.go)** - File utilities and validation demo
- **[pagination/main.go](./examples/pagination/main.go)** - Pagination patterns and iterators
- **[upload/main.go](./examples/upload/main.go)** - File upload with progress tracking

Run an example:

```bash
cd examples/upload
go run main.go
```

## 📖 API Reference

### Client Methods

#### Disk Information

- `GetDisk(ctx) (*Disk, error)` - Get disk information

#### File & Folder Operations

- `GetMetadata(ctx, path) (*Resource, *ErrorResponse)` - Get resource metadata
- `CreateFolder(ctx, path) (*Link, error)` - Create a folder
- `CopyResource(ctx, from, to, overwrite) (*Link, error)` - Copy resource
- `MoveResource(ctx, from, to, overwrite) (*Link, error)` - Move resource
- `DeleteResource(ctx, path, permanently) error` - Delete resource

#### Upload & Download

- `UploadFileFromPath(ctx, local, remote, options) (*Resource, error)` - Upload file
- `UploadLargeFileFromPath(ctx, local, remote, options) (*Resource, error)` - Upload large file
- `UploadFile(ctx, reader, remote, options) (*Resource, error)` - Upload from reader
- `DownloadFile(ctx, remote, local) error` - Download file
- `GetDownloadURL(ctx, path) (*Link, *ErrorResponse)` - Get download URL

#### Public Resources

- `PublishResource(ctx, path) (*Link, error)` - Publish resource
- `UnpublishResource(ctx, path) (*Link, error)` - Unpublish resource
- `GetMetadataForPublicResource(ctx, publicKey) (*PublicResource, *ErrorResponse)`
- `GetDownloadURLForPublicResource(ctx, publicKey) (*Link, *ErrorResponse)`
- `GetPublicResources(ctx) (*PublicResourcesList, *ErrorResponse)`
- `GetPublicResourcesPaged(ctx, options) (*PagedPublicResourcesList, *ErrorResponse)`
- `GetPublicResourcesIterator(options) *OffsetPaginationIterator[*PublicResourcesList]`

#### Trash Operations

- `ListTrashResources(ctx, path, limit, offset) (*TrashResourceList, error)`
- `RestoreFromTrash(ctx, path, overwrite, name) (*Link, error)`
- `DeleteFromTrash(ctx, path) (*Link, error)`
- `EmptyTrash(ctx) (*Link, error)`

#### Batch Operations

- `BatchDeleteFiles(ctx, paths, options) (*BatchOperationStatus, error)`
- `BatchCopyFiles(ctx, sourceDestMap, options) (*BatchOperationStatus, error)`
- `BatchMoveFiles(ctx, sourceDestMap, options) (*BatchOperationStatus, error)`

#### Pagination

- `GetSortedFiles(ctx) (*FilesResourceList, *ErrorResponse)`
- `GetSortedFilesWithPagination(ctx, options) (*FilesResourceList, *ErrorResponse)`
- `GetSortedFilesPaged(ctx, options) (*PagedFilesResourceList, *ErrorResponse)`
- `GetSortedFilesIterator(options) *OffsetPaginationIterator[*FilesResourceList]`
- `GetLastUploadedResources(ctx) (*LastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesWithPagination(ctx, options) (*LastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesPaged(ctx, options) (*PagedLastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesIterator(options) *OffsetPaginationIterator[*LastUploadedResourceList]`

#### Operations

- `OperationStatus(ctx, operationID) (any, *http.Response, error)` - Check async operation status

For detailed pagination API documentation, see [PAGINATION.md](./PAGINATION.md).

## 🧪 Testing

The library includes comprehensive unit tests.

```bash
# Run all tests
go test -v

# Run with coverage
go test -v -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific tests
go test -v -run TestPagination
go test -v -run TestBatch
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/ilyabrin/disk.git
cd disk
```

1. Install dependencies:

```bash
go mod download
```

1. Run tests:

```bash
go test -v
```

### Guidelines

- Write tests for new features
- Follow Go best practices and idioms
- Update documentation for API changes
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🔗 Links

- [Yandex Disk REST API Documentation](https://yandex.ru/dev/disk/rest/)
- [Yandex OAuth](https://oauth.yandex.ru/)
- [GitHub Repository](https://github.com/ilyabrin/disk)

## 📝 Changelog

### Recent Updates

- ✅ Comprehensive pagination support with multiple strategies
- ✅ Batch operations for efficient multi-file processing
- ✅ Enhanced upload functionality with progress tracking
- ✅ Improved error handling and logging
- ✅ Full context.Context support
- ✅ Security enhancements with path validation

## 💬 Support

If you have questions or need help:

- Open an [issue](https://github.com/ilyabrin/disk/issues)
- Check existing [examples](./examples/)
- Read the [pagination documentation](./PAGINATION.md)

## ⭐ Acknowledgments

Built with ❤️ for the Go community. If you find this library helpful, please consider giving it a star on GitHub!

---

**Note:** This library is not officially affiliated with Yandex. It's a community-maintained client library for the Yandex Disk API.
