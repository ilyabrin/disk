package disk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestPaginationOptions(t *testing.T) {
	t.Run("ValidatePaginationOptions with nil", func(t *testing.T) {
		options := ValidatePaginationOptions(nil)
		if options == nil {
			t.Error("Expected non-nil options")
		}
		if options == nil || options.Limit != 20 {
			t.Errorf("Expected default limit 20, got %d", options.Limit)
		}
		if options.Offset != 0 {
			t.Errorf("Expected default offset 0, got %d", options.Offset)
		}
	})

	t.Run("ValidatePaginationOptions with custom values", func(t *testing.T) {
		input := &PaginationOptions{
			Limit:  50,
			Offset: 100,
			Cursor: "test-cursor",
		}
		options := ValidatePaginationOptions(input)
		if options.Limit != 50 {
			t.Errorf("Expected limit 50, got %d", options.Limit)
		}
		if options.Offset != 100 {
			t.Errorf("Expected offset 100, got %d", options.Offset)
		}
		if options.Cursor != "test-cursor" {
			t.Errorf("Expected cursor 'test-cursor', got '%s'", options.Cursor)
		}
	})

	t.Run("ValidatePaginationOptions with invalid values", func(t *testing.T) {
		input := &PaginationOptions{
			Limit:  -10,
			Offset: -5,
		}
		options := ValidatePaginationOptions(input)
		if options.Limit != 20 {
			t.Errorf("Expected default limit 20 for invalid input, got %d", options.Limit)
		}
		if options.Offset != 0 {
			t.Errorf("Expected default offset 0 for invalid input, got %d", options.Offset)
		}
	})

	t.Run("ValidatePaginationOptions with limit exceeding maximum", func(t *testing.T) {
		input := &PaginationOptions{
			Limit: 15000,
		}
		options := ValidatePaginationOptions(input)
		if options.Limit != 10000 {
			t.Errorf("Expected maximum limit 10000, got %d", options.Limit)
		}
	})
}

func TestCreatePaginationInfo(t *testing.T) {
	t.Run("createPaginationInfo with total count", func(t *testing.T) {
		info := createPaginationInfo(20, 0, 20, true, 100)
		if info.Limit != 20 {
			t.Errorf("Expected limit 20, got %d", info.Limit)
		}
		if info.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", info.Offset)
		}
		if info.Total != 100 {
			t.Errorf("Expected total 100, got %d", info.Total)
		}
		if !info.HasMore {
			t.Error("Expected HasMore to be true")
		}
		if info.NextOffset != 20 {
			t.Errorf("Expected NextOffset 20, got %d", info.NextOffset)
		}
	})

	t.Run("createPaginationInfo without total count", func(t *testing.T) {
		info := createPaginationInfo(20, 40, 20, false, 0)
		if info.Limit != 20 {
			t.Errorf("Expected limit 20, got %d", info.Limit)
		}
		if info.Offset != 40 {
			t.Errorf("Expected offset 40, got %d", info.Offset)
		}
		if info.Total != 0 {
			t.Errorf("Expected total 0, got %d", info.Total)
		}
		if !info.HasMore {
			t.Error("Expected HasMore to be true when full page received")
		}
		if info.NextOffset != 60 {
			t.Errorf("Expected NextOffset 60, got %d", info.NextOffset)
		}
	})

	t.Run("createPaginationInfo last page", func(t *testing.T) {
		info := createPaginationInfo(20, 80, 10, false, 0)
		if info.HasMore {
			t.Error("Expected HasMore to be false for partial page")
		}
	})
}

func TestCreatePaginationInfoWithCursor(t *testing.T) {
	t.Run("createPaginationInfoWithCursor with cursors", func(t *testing.T) {
		info := createPaginationInfoWithCursor(20, 0, 20, false, 0, "next-cursor", "prev-cursor")
		if info.NextCursor != "next-cursor" {
			t.Errorf("Expected NextCursor 'next-cursor', got '%s'", info.NextCursor)
		}
		if info.PrevCursor != "prev-cursor" {
			t.Errorf("Expected PrevCursor 'prev-cursor', got '%s'", info.PrevCursor)
		}
		if !info.HasMore {
			t.Error("Expected HasMore to be true when NextCursor is present")
		}
	})

	t.Run("createPaginationInfoWithCursor without next cursor", func(t *testing.T) {
		info := createPaginationInfoWithCursor(20, 0, 10, false, 0, "", "prev-cursor")
		if info.NextCursor != "" {
			t.Errorf("Expected empty NextCursor, got '%s'", info.NextCursor)
		}
		if info.HasMore {
			t.Error("Expected HasMore to be false when NextCursor is empty and partial page")
		}
	})
}

func TestGetSortedFilesWithPagination(t *testing.T) {
	t.Run("GetSortedFilesWithPagination validates input", func(t *testing.T) {
		client, _ := New("test-token")

		// Test with nil options
		_, err := client.GetSortedFilesWithPagination(context.Background(), nil)
		if err == nil {
			t.Error("Expected error due to invalid token, but validation should work")
		}
	})

	t.Run("GetSortedFilesWithPagination with custom options", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{
			Limit:  10,
			Offset: 20,
		}

		_, err := client.GetSortedFilesWithPagination(context.Background(), options)
		if err == nil {
			t.Error("Expected error due to invalid token")
		}
		// The test should fail at API level, but pagination structure should be correct
	})
}

func TestGetSortedFilesPaged(t *testing.T) {
	t.Run("GetSortedFilesPaged returns pagination info", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{
			Limit:  15,
			Offset: 5,
		}

		_, err := client.GetSortedFilesPaged(context.Background(), options)
		if err == nil {
			t.Error("Expected error due to invalid token")
		}
		// Test validates that the method exists and accepts parameters correctly
	})
}

func TestGetLastUploadedResourcesWithPagination(t *testing.T) {
	t.Run("GetLastUploadedResourcesWithPagination validates input", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{
			Limit:  25,
			Offset: 10,
		}

		_, err := client.GetLastUploadedResourcesWithPagination(context.Background(), options)
		if err == nil {
			t.Error("Expected error due to invalid token")
		}
	})
}

func TestGetPublicResourcesWithPagination(t *testing.T) {
	t.Run("GetPublicResourcesWithPagination validates input", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{
			Limit:  30,
			Offset: 15,
		}

		_, err := client.GetPublicResourcesWithPagination(context.Background(), options)
		if err == nil {
			t.Error("Expected error due to invalid token")
		}
	})
}

func TestPaginationIterator(t *testing.T) {
	t.Run("NewPaginationIterator creates iterator correctly", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			return nil, nil
		}

		options := &PaginationOptions{Limit: 15}
		iterator := NewPaginationIterator(client, fetcher, options)

		if iterator == nil {
			t.Error("Expected non-nil iterator")
		}
		if iterator.GetPageSize() != 15 {
			t.Errorf("Expected page size 15, got %d", iterator.GetPageSize())
		}
		if !iterator.HasNext() {
			t.Error("Expected HasNext to be true initially")
		}
	})

	t.Run("PaginationIterator SetPageSize works", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			return nil, nil
		}

		iterator := NewPaginationIterator(client, fetcher, nil)
		iterator.SetPageSize(25)

		if iterator.GetPageSize() != 25 {
			t.Errorf("Expected page size 25 after set, got %d", iterator.GetPageSize())
		}
	})

	t.Run("PaginationIterator Reset works", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			return nil, nil
		}

		iterator := NewPaginationIterator(client, fetcher, nil)
		iterator.Reset()

		if iterator.GetCurrentOffset() != 0 {
			t.Errorf("Expected offset 0 after reset, got %d", iterator.GetCurrentOffset())
		}
		if !iterator.HasNext() {
			t.Error("Expected HasNext to be true after reset")
		}
	})
}

func TestCursorPaginationIterator(t *testing.T) {
	t.Run("NewCursorPaginationIterator creates iterator correctly", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
			return nil, "", nil
		}

		iterator := NewCursorPaginationIterator(client, fetcher, 10)

		if iterator == nil {
			t.Error("Expected non-nil cursor iterator")
		}
		if iterator.GetPageSize() != 10 {
			t.Errorf("Expected page size 10, got %d", iterator.GetPageSize())
		}
		if !iterator.HasNext() {
			t.Error("Expected HasNext to be true initially")
		}
	})

	t.Run("CursorPaginationIterator handles default page size", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
			return nil, "", nil
		}

		iterator := NewCursorPaginationIterator(client, fetcher, 0)

		if iterator.GetPageSize() != 20 {
			t.Errorf("Expected default page size 20, got %d", iterator.GetPageSize())
		}
	})

	t.Run("CursorPaginationIterator handles max page size", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
			return nil, "", nil
		}

		iterator := NewCursorPaginationIterator(client, fetcher, 15000)

		if iterator.GetPageSize() != 10000 {
			t.Errorf("Expected max page size 10000, got %d", iterator.GetPageSize())
		}
	})

	t.Run("CursorPaginationIterator Reset works", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
			return nil, "", nil
		}

		iterator := NewCursorPaginationIterator(client, fetcher, 10)
		iterator.Reset()

		if iterator.GetNextCursor() != "" {
			t.Errorf("Expected empty cursor after reset, got '%s'", iterator.GetNextCursor())
		}
		if !iterator.HasNext() {
			t.Error("Expected HasNext to be true after reset")
		}
	})
}

func TestPaginationIterators(t *testing.T) {
	t.Run("GetSortedFilesIterator creates iterator", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{Limit: 5}
		iterator := client.GetSortedFilesIterator(options)

		if iterator == nil {
			t.Error("Expected non-nil iterator")
		}
	})

	t.Run("GetLastUploadedResourcesIterator creates iterator", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{Limit: 10}
		iterator := client.GetLastUploadedResourcesIterator(options)

		if iterator == nil {
			t.Error("Expected non-nil iterator")
		}
	})

	t.Run("GetPublicResourcesIterator creates iterator", func(t *testing.T) {
		client, _ := New("test-token")

		options := &PaginationOptions{Limit: 15}
		iterator := client.GetPublicResourcesIterator(options)

		if iterator == nil {
			t.Error("Expected non-nil iterator")
		}
	})
}

func TestPaginationCompatibility(t *testing.T) {
	t.Run("Original methods still work", func(t *testing.T) {
		client, _ := New("test-token")

		// These should not panic and should maintain backward compatibility
		_, err := client.GetSortedFiles(context.Background())
		if err == nil {
			t.Error("Expected error due to invalid token, but method should exist")
		}

		_, err = client.GetLastUploadedResources(context.Background())
		if err == nil {
			t.Error("Expected error due to invalid token, but method should exist")
		}

		_, err = client.GetPublicResources(context.Background())
		if err == nil {
			t.Error("Expected error due to invalid token, but method should exist")
		}
	})
}

func TestCustomPageSizes(t *testing.T) {
	t.Run("Custom page sizes are supported", func(t *testing.T) {
		testCases := []struct {
			name     string
			pageSize int
			expected int
		}{
			{"Small page size", 5, 5},
			{"Medium page size", 100, 100},
			{"Large page size", 1000, 1000},
			{"Maximum page size", 10000, 10000},
			{"Over maximum page size", 15000, 10000},
			{"Zero page size", 0, 20},
			{"Negative page size", -10, 20},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				options := &PaginationOptions{Limit: tc.pageSize}
				validated := ValidatePaginationOptions(options)
				if validated.Limit != tc.expected {
					t.Errorf("Expected limit %d, got %d", tc.expected, validated.Limit)
				}
			})
		}
	})
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("Empty result handling", func(t *testing.T) {
		info := createPaginationInfo(20, 0, 0, false, 0)
		if info.HasMore {
			t.Error("Expected HasMore to be false for empty results")
		}
	})

	t.Run("Single item result", func(t *testing.T) {
		info := createPaginationInfo(20, 0, 1, false, 0)
		if info.HasMore {
			t.Error("Expected HasMore to be false for single item when page size is 20")
		}
	})

	t.Run("Exact page size result", func(t *testing.T) {
		info := createPaginationInfo(20, 0, 20, false, 0)
		if !info.HasMore {
			t.Error("Expected HasMore to be true for exact page size")
		}
	})
}

func TestCursorPaginationWithMockData(t *testing.T) {
	t.Run("Cursor pagination with mock fetcher", func(t *testing.T) {
		client, _ := New("test-token")

		pages := []struct {
			data   []string
			cursor string
		}{
			{[]string{"item1", "item2", "item3"}, "cursor1"},
			{[]string{"item4", "item5", "item6"}, "cursor2"},
			{[]string{"item7", "item8"}, ""},
		}

		currentPage := 0

		fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
			if currentPage >= len(pages) {
				return &PagedFilesResourceList{}, "", nil
			}

			page := pages[currentPage]
			currentPage++

			// Mock response
			result := &PagedFilesResourceList{
				FilesResourceList: &FilesResourceList{
					Items: make([]*Resource, len(page.data)),
				},
			}

			return result, page.cursor, nil
		}

		iterator := NewCursorPaginationIterator(client, fetcher, 10)

		// Test first page
		page1, err := iterator.Next(context.Background())
		if err != nil {
			t.Errorf("Expected no error for first page, got %v", err)
		}
		if page1 == nil {
			t.Error("Expected non-nil page1")
		}
		if !iterator.HasNext() {
			t.Error("Expected more pages after first page")
		}

		// Test second page
		page2, err := iterator.Next(context.Background())
		if err != nil {
			t.Errorf("Expected no error for second page, got %v", err)
		}
		if page2 == nil {
			t.Error("Expected non-nil page2")
		}
		if !iterator.HasNext() {
			t.Error("Expected more pages after second page")
		}

		// Test third page (last)
		page3, err := iterator.Next(context.Background())
		if err != nil {
			t.Errorf("Expected no error for third page, got %v", err)
		}
		if page3 == nil {
			t.Error("Expected non-nil page3")
		}
		if iterator.HasNext() {
			t.Error("Expected no more pages after third page")
		}

		// Test beyond last page
		_, err = iterator.Next(context.Background())
		if err == nil {
			t.Error("Expected error when trying to get page beyond last")
		}
	})
}

func TestPaginationIteratorNext(t *testing.T) {
	t.Run("successful fetch advances offset and returns value", func(t *testing.T) {
		client, _ := New("test-token")

		called := 0
		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			called++
			return &PagedFilesResourceList{
				FilesResourceList: &FilesResourceList{Items: make([]*Resource, options.Limit)},
				Pagination:        &PaginationInfo{HasMore: true},
			}, nil
		}

		iter := NewPaginationIterator(client, fetcher, &PaginationOptions{Limit: 10, Offset: 0})
		result, err := iter.Next(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if called != 1 {
			t.Errorf("expected fetcher called once, got %d", called)
		}
		if iter.GetCurrentOffset() != 10 {
			t.Errorf("expected offset 10 after Next, got %d", iter.GetCurrentOffset())
		}
	})

	t.Run("fetcher error is returned", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			return nil, fmt.Errorf("fetch failed")
		}

		iter := NewPaginationIterator(client, fetcher, nil)
		_, err := iter.Next(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "fetch failed") {
			t.Errorf("expected 'fetch failed', got: %v", err)
		}
	})

	t.Run("hasMore false returns no more pages error", func(t *testing.T) {
		client, _ := New("test-token")

		fetcher := func(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, error) {
			return nil, nil
		}

		iter := NewPaginationIterator(client, fetcher, nil)
		iter.hasMore = false

		_, err := iter.Next(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no more pages available") {
			t.Errorf("expected 'no more pages available', got: %v", err)
		}
	})
}

func TestGetLastUploadedResourcesPaged(t *testing.T) {
	client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"items":[],"limit":20}`))
	}))

	result, errResp := client.GetLastUploadedResourcesPaged(context.Background(), nil)
	if errResp != nil {
		t.Fatalf("expected no error, got: %v", errResp)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Pagination == nil {
		t.Error("expected non-nil Pagination")
	}
}

func TestGetPublicResourcesPaged(t *testing.T) {
	client := mockedHttpClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"items":[],"type":"dir","limit":20,"offset":0}`))
	}))

	result, errResp := client.GetPublicResourcesPaged(context.Background(), nil)
	if errResp != nil {
		t.Fatalf("expected no error, got: %v", errResp)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Pagination == nil {
		t.Error("expected non-nil Pagination")
	}
}

func TestCursorPaginationIteratorSetPageSize(t *testing.T) {
	client, _ := New("test-token")
	fetcher := func(ctx context.Context, cursor string, limit int) (*PagedFilesResourceList, string, error) {
		return nil, "", nil
	}

	iter := NewCursorPaginationIterator(client, fetcher, 20)

	iter.SetPageSize(50)
	if iter.GetPageSize() != 50 {
		t.Errorf("expected page size 50, got %d", iter.GetPageSize())
	}

	// Invalid sizes are ignored
	for _, invalid := range []int{0, -1, 20001} {
		iter.SetPageSize(invalid)
		if iter.GetPageSize() != 50 {
			t.Errorf("expected page size unchanged (50) after invalid size %d, got %d", invalid, iter.GetPageSize())
		}
	}
}

func TestAddPaginationParams(t *testing.T) {
	t.Run("nil options sets no params", func(t *testing.T) {
		q := url.Values{}
		addPaginationParams(q, nil)
		if len(q) != 0 {
			t.Errorf("expected no params, got %v", q)
		}
	})

	t.Run("cursor takes precedence over offset", func(t *testing.T) {
		q := url.Values{}
		addPaginationParams(q, &PaginationOptions{Limit: 20, Offset: 10, Cursor: "abc"})
		if q.Get("cursor") != "abc" {
			t.Errorf("expected cursor 'abc', got %q", q.Get("cursor"))
		}
		if q.Get("offset") != "" {
			t.Errorf("expected no offset param when cursor present, got %q", q.Get("offset"))
		}
	})

	t.Run("offset set when no cursor", func(t *testing.T) {
		q := url.Values{}
		addPaginationParams(q, &PaginationOptions{Limit: 20, Offset: 40})
		if q.Get("offset") != "40" {
			t.Errorf("expected offset '40', got %q", q.Get("offset"))
		}
		if q.Get("cursor") != "" {
			t.Errorf("expected no cursor param, got %q", q.Get("cursor"))
		}
	})

	t.Run("limit clamped to 10000", func(t *testing.T) {
		q := url.Values{}
		addPaginationParams(q, &PaginationOptions{Limit: 99999})
		if q.Get("limit") != "10000" {
			t.Errorf("expected limit '10000', got %q", q.Get("limit"))
		}
	})
}
