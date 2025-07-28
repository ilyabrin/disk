package disk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PaginationOptions contains options for paginated requests
type PaginationOptions struct {
	Limit  int    // Maximum number of items to return (default: 20, max: 10000)
	Offset int    // Number of items to skip from the beginning (default: 0)
	Cursor string // Cursor for cursor-based pagination (optional)
}

// PaginationInfo contains pagination metadata from the response
type PaginationInfo struct {
	Limit      int    // Number of items requested
	Offset     int    // Number of items skipped
	Total      int    // Total number of items available (when available)
	HasMore    bool   // Whether there are more items available
	NextOffset int    // Offset for the next page
	NextCursor string // Cursor for the next page (cursor-based pagination)
	PrevCursor string // Cursor for the previous page (cursor-based pagination)
}

// PagedFilesResourceList contains paginated files with pagination info
type PagedFilesResourceList struct {
	*FilesResourceList
	Pagination *PaginationInfo `json:"pagination"`
}

// PagedLastUploadedResourceList contains paginated last uploaded resources with pagination info
type PagedLastUploadedResourceList struct {
	*LastUploadedResourceList
	Pagination *PaginationInfo `json:"pagination"`
}

// PagedPublicResourcesList contains paginated public resources with pagination info
type PagedPublicResourcesList struct {
	*PublicResourcesList
	Pagination *PaginationInfo `json:"pagination"`
}

// PaginationIterator provides an iterator interface for paginated results
type PaginationIterator[T any] struct {
	client     *Client
	fetcher    func(ctx context.Context, options *PaginationOptions) (T, error)
	options    *PaginationOptions
	hasMore    bool
	totalItems int
}

// NewPaginationIterator creates a new pagination iterator
func NewPaginationIterator[T any](
	client *Client,
	fetcher func(ctx context.Context, options *PaginationOptions) (T, error),
	options *PaginationOptions,
) *PaginationIterator[T] {
	if options == nil {
		options = &PaginationOptions{Limit: 20, Offset: 0}
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}
	if options.Limit > 10000 {
		options.Limit = 10000
	}
	if options.Offset < 0 {
		options.Offset = 0
	}

	return &PaginationIterator[T]{
		client:  client,
		fetcher: fetcher,
		options: options,
		hasMore: true,
	}
}

// Next fetches the next page of results
func (p *PaginationIterator[T]) Next(ctx context.Context) (T, error) {
	if !p.hasMore {
		var zero T
		return zero, fmt.Errorf("no more pages available")
	}

	result, err := p.fetcher(ctx, p.options)
	if err != nil {
		var zero T
		return zero, err
	}

	// Update iterator state for next page
	p.options.Offset += p.options.Limit

	// Determine if there are more pages (this logic will be customized per type)
	// For now, we assume there are no more pages if we got fewer items than requested
	// This will be refined in the specific implementations

	return result, nil
}

// HasNext returns true if there are more pages available
func (p *PaginationIterator[T]) HasNext() bool {
	return p.hasMore
}

// Reset resets the iterator to the beginning
func (p *PaginationIterator[T]) Reset() {
	p.options.Offset = 0
	p.hasMore = true
}

// SetPageSize sets the page size for future requests
func (p *PaginationIterator[T]) SetPageSize(size int) {
	if size > 0 && size <= 10000 {
		p.options.Limit = size
	}
}

// GetPageSize returns the current page size
func (p *PaginationIterator[T]) GetPageSize() int {
	return p.options.Limit
}

// GetCurrentOffset returns the current offset
func (p *PaginationIterator[T]) GetCurrentOffset() int {
	return p.options.Offset
}

// addPaginationParams adds pagination parameters to URL query values
func addPaginationParams(query url.Values, options *PaginationOptions) {
	if options == nil {
		return
	}

	if options.Limit > 0 {
		if options.Limit > 10000 {
			options.Limit = 10000
		}
		query.Set("limit", strconv.Itoa(options.Limit))
	}

	// Prefer cursor over offset if both are provided
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	} else if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
}

// createPaginationInfo creates pagination info from response data
func createPaginationInfo(limit, offset int, itemCount int, hasTotal bool, total int) *PaginationInfo {
	info := &PaginationInfo{
		Limit:  limit,
		Offset: offset,
	}

	if hasTotal {
		info.Total = total
		info.HasMore = offset+itemCount < total
		if info.HasMore {
			info.NextOffset = offset + limit
		}
	} else {
		// If we don't have total count, assume there are more if we got a full page
		info.HasMore = itemCount >= limit
		if info.HasMore {
			info.NextOffset = offset + limit
		}
	}

	return info
}

// createPaginationInfoWithCursor creates pagination info with cursor support
func createPaginationInfoWithCursor(limit, offset int, itemCount int, hasTotal bool, total int, nextCursor, prevCursor string) *PaginationInfo {
	info := createPaginationInfo(limit, offset, itemCount, hasTotal, total)
	info.NextCursor = nextCursor
	info.PrevCursor = prevCursor

	// If we have cursors, we can determine HasMore from NextCursor presence
	if nextCursor != "" {
		info.HasMore = true
	} else if nextCursor == "" && itemCount < limit {
		info.HasMore = false
	}

	return info
}

// CursorPaginationIterator provides cursor-based pagination iterator
type CursorPaginationIterator[T any] struct {
	client      *Client
	fetcher     func(ctx context.Context, cursor string, limit int) (T, string, error)
	limit       int
	currentPage T
	nextCursor  string
	hasMore     bool
	initialized bool
}

// NewCursorPaginationIterator creates a new cursor-based pagination iterator
func NewCursorPaginationIterator[T any](
	client *Client,
	fetcher func(ctx context.Context, cursor string, limit int) (T, string, error),
	limit int,
) *CursorPaginationIterator[T] {
	if limit <= 0 {
		limit = 20
	}
	if limit > 10000 {
		limit = 10000
	}

	return &CursorPaginationIterator[T]{
		client:      client,
		fetcher:     fetcher,
		limit:       limit,
		hasMore:     true,
		initialized: false,
	}
}

// Next fetches the next page using cursor-based pagination
func (c *CursorPaginationIterator[T]) Next(ctx context.Context) (T, error) {
	if !c.hasMore {
		var zero T
		return zero, fmt.Errorf("no more pages available")
	}

	cursor := ""
	if c.initialized {
		cursor = c.nextCursor
	}

	result, nextCursor, err := c.fetcher(ctx, cursor, c.limit)
	if err != nil {
		var zero T
		return zero, err
	}

	c.currentPage = result
	c.nextCursor = nextCursor
	c.hasMore = nextCursor != ""
	c.initialized = true

	return result, nil
}

// HasNext returns true if there are more pages available
func (c *CursorPaginationIterator[T]) HasNext() bool {
	return c.hasMore
}

// Reset resets the iterator to the beginning
func (c *CursorPaginationIterator[T]) Reset() {
	c.nextCursor = ""
	c.hasMore = true
	c.initialized = false
}

// GetPageSize returns the current page size
func (c *CursorPaginationIterator[T]) GetPageSize() int {
	return c.limit
}

// SetPageSize sets the page size for future requests
func (c *CursorPaginationIterator[T]) SetPageSize(size int) {
	if size > 0 && size <= 10000 {
		c.limit = size
	}
}

// GetNextCursor returns the cursor for the next page
func (c *CursorPaginationIterator[T]) GetNextCursor() string {
	return c.nextCursor
}

// ValidatePaginationOptions validates and normalizes pagination options
func ValidatePaginationOptions(options *PaginationOptions) *PaginationOptions {
	if options == nil {
		return &PaginationOptions{Limit: 20, Offset: 0}
	}

	normalized := &PaginationOptions{
		Limit:  options.Limit,
		Offset: options.Offset,
		Cursor: options.Cursor,
	}

	if normalized.Limit <= 0 {
		normalized.Limit = 20
	}
	if normalized.Limit > 10000 {
		normalized.Limit = 10000
	}
	if normalized.Offset < 0 {
		normalized.Offset = 0
	}

	return normalized
}
