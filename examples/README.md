# Yandex Disk Upload Examples

This directory contains examples demonstrating the file upload functionality implemented in section 2.2.

## Examples

### 1. `demo.go` - Utility Functions Demo

A demonstration of utility functions that work without requiring a Yandex Disk token:

```bash
go run demo.go test_file.txt
```

**Features demonstrated:**

- File size detection and formatting
- Path validation for Yandex Disk
- Upload method recommendations based on file size
- File size formatting examples

### 2. `upload_example.go` - Full Upload Example

A complete example showing how to upload files to Yandex Disk with progress tracking:

```bash
# Set your OAuth token
export YANDEX_DISK_TOKEN="your_token_here"

# Upload a file
go run upload_example.go test_file.txt /uploaded/test_file.txt
```

**Features demonstrated:**

- File upload with progress tracking
- Automatic selection of upload method based on file size
- Progress callback with formatted file sizes
- Error handling and validation

## Getting a Yandex Disk Token

1. Go to [Yandex Disk API Polygon](https://yandex.ru/dev/disk/poligon/)
2. Click "Get OAuth token"
3. Authorize the application
4. Copy the token and set it as an environment variable:

   ```bash
   export YANDEX_DISK_TOKEN="your_token_here"
   ```

## Upload Methods Available

### Basic Upload

```go
resource, err := client.UploadFileFromPath(ctx, localPath, remotePath, options)
```

### Upload with Progress

```go
resource, err := client.UploadFileFromPathWithProgress(ctx, localPath, remotePath, overwrite, progressCallback)
```

### Large File Upload

```go
resource, err := client.UploadLargeFileFromPath(ctx, localPath, remotePath, chunkSizeMB, progressCallback)
```

## Upload Options

```go
options := &disk.UploadOptions{
    Overwrite:        true,           // Overwrite existing files
    Progress:         progressFunc,   // Progress callback function
    ChunkSize:        10 * 1024 * 1024, // 10MB chunks for large files
    ValidateChecksum: false,          // Future: validate checksums
}
```

## Progress Callback

```go
progressCallback := func(progress disk.UploadProgress) {
    fmt.Printf("Uploading... %.1f%% (%s / %s)\n", 
        progress.Percentage,
        disk.FormatFileSize(progress.BytesUploaded),
        disk.FormatFileSize(progress.TotalBytes))
}
```

## File Utilities

```go
// Get file size
size, err := disk.GetFileSize(filePath)

// Format file size for display
formatted := disk.FormatFileSize(size)

// Validate path for Yandex Disk
err := disk.ValidateFilePath(remotePath)

// Detect MIME type
mimeType, err := client.DetectMimeType(filePath)
```
