# disk

Yandex.Disk API client | [REST API](https://yandex.ru/dev/disk/rest/)

![GitHub](https://img.shields.io/github/license/ilyabrin/disk)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/ilyabrin/disk)
[![Coverage Status](https://coveralls.io/repos/github/ilyabrin/disk/badge.svg?branch=release)](https://coveralls.io/github/ilyabrin/disk?branch=release)
![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/ilyabrin/disk)
<!-- ![GitHub All Releases](https://img.shields.io/github/downloads/ilyabrin/disk/total) -->

## Install

```sh
go get github.com/ilyabrin/disk
```

## Using

```go
package main

import (
  "context"
  "log"

  "github.com/ilyabrin/disk"
)

func main() {
  ctx := context.Background()

  client := disk.New("YOUR_ACCESS_TOKEN")

  disk, errorResponse := client.Disk.Info(ctx, nil)
  if errorResponse != nil {
    log.Fatal(errorResponse)
  }

  log.Println(disk)
}

```

## Available methods

[Full documentation available here](https://pkg.go.dev/github.com/ilyabrin/disk#section-documentation)

### Disk

```go
// Disk information
client.Disk.Info()
```

### Resources

```go
// Remove resource
client.Resources.Delete(ctx, "path_to_file", false, nil)

// Get meta information
client.Resources.Meta(ctx, "path_to_file", nil)

// Update information
client.Resources.UpdateMeta(ctx, "path_to_file", newMeta)

<details>
  <summary>Example</summary>
    ```go
    newMeta := &disk.Metadata{
      "custom_properties": {
       "key": "value",
       "foo": "bar",
       "platform": "linux",
      },
    }

    client.Resources.UpdateMeta(ctx, "path_to_file", newMeta)
    ```
</details>

// Create directory
client.Resources.CreateDir(ctx, "path_to_file", nil)

// Copy resource
client.Resources.Copy(ctx, "path_to_file", "copy_here", nil)

// Get download link for resource
client.Resources.DownloadURL(ctx, "name_for_uploaded_file", nil)

// Get list of sorted files
client.Resources.GetSortedFiles(ctx, nil)

// List of last uploaded files
client.Resources.ListLastUploaded(ctx, nil)

// Move resource or rename it
client.Resources.Move(ctx, "path_to_file", "move_here", nil)

// List all public links
client.Resources.ListPublic(ctx, nil)

// Publish resource
client.Resources.Publish(ctx, "path_to_file", nil)

// Unpublish resource
client.Resources.Unpublish(ctx, "path_to_file", nil)

// Get link for ipload new file to Disk
client.Resources.GetUploadLink(ctx, "path_to_file")

// Upload new file to Disk
client.Resources.Upload(ctx, "local_file_path", "link_from_GetUploadLink", nil)
```

### Public

```go
// Get metadata for public resource
client.Public.Meta(ctx, "full_url_to_file", nil)

// Get link for download public resource
client.Public.DownloadURL(ctx, "full_url_to_file", nil)

// Save public file to Disk
client.Public.Save(ctx, "full_url_to_file", nil)
```

### Trash

```go
// Delete file from trash
client.Trash.Delete(ctx, "path_to_file", nil)

// Restore file from trash
client.Trash.Restore(ctx, "path_to_file", nil)

// List all resources OR all about resource in Trash
client.Trash.List(ctx, "/", nil)
```

### Operations

```go
// Get operation status
client.Operations.Status(ctx, "operation_id_string", nil)
```

# License

Licensed under the MIT License, Copyright © 2020-present Ilya Brin
