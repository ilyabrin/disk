# Pagination Support

This document describes the comprehensive pagination support implemented in the Yandex Disk Go client library.

## Overview

The pagination system provides multiple ways to handle large result sets from the Yandex Disk API:

1. **Basic Pagination** - Simple offset/limit-based pagination
2. **Enhanced Pagination** - Pagination with metadata and status information
3. **Iterator Pattern** - Convenient iteration over paginated results
4. **Cursor-based Pagination** - Future-ready cursor support

## API Methods with Pagination Support

### GetSortedFiles

Get a sorted list of files with pagination support.

**Basic Usage:**
```go
// Get first 20 files (default)
files, err := client.GetSortedFiles(ctx)

// Get with custom pagination
options := &disk.PaginationOptions{
    Limit:  10,
    Offset: 20,
}
files, err := client.GetSortedFilesWithPagination(ctx, options)
```

**Enhanced Pagination:**
```go
options := &disk.PaginationOptions{Limit: 15}
pagedFiles, err := client.GetSortedFilesPaged(ctx, options)

if err == nil {
    fmt.Printf("Files: %d\n", len(pagedFiles.Items))
    fmt.Printf("HasMore: %t\n", pagedFiles.Pagination.HasMore)
    if pagedFiles.Pagination.HasMore {
        fmt.Printf("NextOffset: %d\n", pagedFiles.Pagination.NextOffset)
    }
}
```

**Iterator Pattern:**
```go
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 10})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        break
    }
    
    for _, file := range page.FilesResourceList.Items {
        fmt.Printf("File: %s\n", file.Name)
    }
}
```

### GetLastUploadedResources

Get recently uploaded files with pagination.

```go
// Basic pagination
files, err := client.GetLastUploadedResources(ctx)

// With custom options
options := &disk.PaginationOptions{Limit: 5}
files, err := client.GetLastUploadedResourcesWithPagination(ctx, options)

// Enhanced with pagination info
pagedFiles, err := client.GetLastUploadedResourcesPaged(ctx, options)

// Iterator pattern
iterator := client.GetLastUploadedResourcesIterator(options)
```

### GetPublicResources

Get public resources with pagination.

```go
// Basic pagination
resources, err := client.GetPublicResources(ctx)

// With custom options
options := &disk.PaginationOptions{Limit: 10}
resources, err := client.GetPublicResourcesWithPagination(ctx, options)

// Enhanced with pagination info
pagedResources, err := client.GetPublicResourcesPaged(ctx, options)

// Iterator pattern
iterator := client.GetPublicResourcesIterator(options)
```

## Pagination Options

### PaginationOptions Structure

```go
type PaginationOptions struct {
    Limit  int    // Maximum number of items to return (default: 20, max: 10000)
    Offset int    // Number of items to skip from the beginning (default: 0)
    Cursor string // Cursor for cursor-based pagination (optional)
}
```

### Default Values

- **Limit**: 20 items per page
- **Offset**: 0 (start from beginning)
- **Maximum Limit**: 10000 items per page

### Validation

All pagination options are automatically validated:
- Negative or zero limits default to 20
- Limits exceeding 10000 are capped at 10000
- Negative offsets are set to 0

## Pagination Information

### PaginationInfo Structure

```go
type PaginationInfo struct {
    Limit      int    // Number of items requested
    Offset     int    // Number of items skipped
    Total      int    // Total number of items available (when available)
    HasMore    bool   // Whether there are more items available
    NextOffset int    // Offset for the next page
    NextCursor string // Cursor for the next page (cursor-based pagination)
    PrevCursor string // Cursor for the previous page (cursor-based pagination)
}
```

## Iterator Patterns

### Basic Iterator

```go
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 25})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        log.Printf("Error: %v", err)
        break
    }
    
    // Process page.FilesResourceList.Items
    for _, file := range page.FilesResourceList.Items {
        fmt.Printf("Processing: %s\n", file.Name)
    }
    
    // Optional: Add delay to respect rate limits
    time.Sleep(200 * time.Millisecond)
}
```

### Iterator Management

```go
iterator := client.GetSortedFilesIterator(nil)

// Change page size
iterator.SetPageSize(50)

// Check current settings
pageSize := iterator.GetPageSize()
currentOffset := iterator.GetCurrentOffset()

// Reset to beginning
iterator.Reset()
```

### Cursor-based Iterator

```go
fetcher := func(ctx context.Context, cursor string, limit int) (*disk.PagedFilesResourceList, string, error) {
    // Custom fetcher implementation
    // Return: (data, nextCursor, error)
}

cursorIterator := disk.NewCursorPaginationIterator(client, fetcher, 20)

for cursorIterator.HasNext() {
    page, err := cursorIterator.Next(ctx)
    if err != nil {
        break
    }
    // Process page
}
```

## Advanced Usage Examples

### Collecting All Results

```go
func collectAllFiles(client *disk.Client, ctx context.Context) ([]*disk.Resource, error) {
    var allFiles []*disk.Resource
    
    iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 100})
    
    for iterator.HasNext() {
        page, err := iterator.Next(ctx)
        if err != nil {
            return nil, err
        }
        
        allFiles = append(allFiles, page.FilesResourceList.Items...)
        
        // Respect rate limits
        time.Sleep(100 * time.Millisecond)
    }
    
    return allFiles, nil
}
```

### Searching with Pagination

```go
func searchFiles(client *disk.Client, ctx context.Context, pattern string) ([]*disk.Resource, error) {
    var matches []*disk.Resource
    
    iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 50})
    
    for iterator.HasNext() {
        page, err := iterator.Next(ctx)
        if err != nil {
            return nil, err
        }
        
        for _, file := range page.FilesResourceList.Items {
            if strings.Contains(strings.ToLower(file.Name), strings.ToLower(pattern)) {
                matches = append(matches, file)
            }
        }
        
        time.Sleep(200 * time.Millisecond)
    }
    
    return matches, nil
}
```

### Custom Page Sizes

```go
// Different page sizes for different use cases
smallPages := &disk.PaginationOptions{Limit: 5}   // For UI display
mediumPages := &disk.PaginationOptions{Limit: 50}  // For processing
largePages := &disk.PaginationOptions{Limit: 1000} // For bulk operations

// Use with any paginated method
files1, _ := client.GetSortedFilesWithPagination(ctx, smallPages)
files2, _ := client.GetLastUploadedResourcesWithPagination(ctx, mediumPages)
files3, _ := client.GetPublicResourcesWithPagination(ctx, largePages)
```

## Best Practices

### Rate Limiting

Always add delays between API calls to respect rate limits:

```go
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 20})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        break
    }
    
    // Process page...
    
    // Add delay between requests
    time.Sleep(200 * time.Millisecond)
}
```

### Error Handling

```go
iterator := client.GetSortedFilesIterator(nil)

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        log.Printf("Error fetching page: %v", err)
        
        // Decide whether to continue or abort
        if isTemporaryError(err) {
            time.Sleep(1 * time.Second)
            continue
        } else {
            break
        }
    }
    
    // Process page...
}
```

### Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 10})

for iterator.HasNext() {
    select {
    case <-ctx.Done():
        log.Printf("Operation cancelled: %v", ctx.Err())
        return
    default:
        page, err := iterator.Next(ctx)
        if err != nil {
            break
        }
        // Process page...
    }
}
```

### Memory Management

For large datasets, process pages individually rather than collecting all results:

```go
// Good: Process each page individually
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 100})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        break
    }
    
    // Process and discard page
    processFiles(page.FilesResourceList.Items)
    // page goes out of scope and can be garbage collected
}

// Avoid: Collecting all results in memory for large datasets
// var allFiles []*disk.Resource // This could consume too much memory
```

## Migration from Non-Paginated Methods

The original methods are preserved for backward compatibility:

```go
// Old way (still works)
files, err := client.GetSortedFiles(ctx)

// New way with explicit pagination
files, err := client.GetSortedFilesWithPagination(ctx, &disk.PaginationOptions{Limit: 20})

// Enhanced way with pagination info
pagedFiles, err := client.GetSortedFilesPaged(ctx, &disk.PaginationOptions{Limit: 20})
```

## Future Enhancements

The pagination system is designed to support future API enhancements:

1. **Cursor-based Pagination** - Ready for when the API supports cursors
2. **Total Count Support** - Will utilize total counts when available
3. **Sorting Options** - Can be extended to support different sort orders
4. **Filtering** - Framework ready for server-side filtering

## Error Handling

All pagination methods return appropriate error types:

```go
files, errResp := client.GetSortedFilesWithPagination(ctx, options)
if errResp != nil {
    switch errResp.Error {
    case "UnauthorizedError":
        // Handle authentication issues
    case "LimitExceededError":
        // Handle rate limiting
    default:
        // Handle other errors
    }
}
```

## Performance Considerations

1. **Page Size**: Balance between fewer requests (larger pages) and memory usage
2. **Rate Limits**: Always include delays between requests
3. **Context Timeouts**: Set appropriate timeouts for large operations
4. **Error Retry**: Implement retry logic for temporary failures
5. **Memory Usage**: Process pages individually for large datasets

## Testing

Comprehensive tests are provided for all pagination functionality:

```bash
# Run pagination-specific tests
go test -v -run TestPagination

# Run all tests to ensure compatibility
go test -v
```

## Examples

See `examples/pagination_example.go` for a complete working example demonstrating all pagination features.