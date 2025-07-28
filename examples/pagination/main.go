package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyabrin/disk"
)

func main() {
	token := os.Getenv("YANDEX_DISK_TOKEN")
	if token == "" {
		log.Fatal("Please set YANDEX_DISK_TOKEN environment variable")
	}

	client, err := disk.New(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	fmt.Println("=== Yandex Disk Pagination Examples ===")

	// Example 1: Basic pagination with offset/limit
	fmt.Println("1. Basic Pagination with GetSortedFiles")
	fmt.Println("--------------------------------------")
	
	options := &disk.PaginationOptions{
		Limit:  5,  // Get 5 files per page
		Offset: 0,  // Start from beginning
	}

	files, errResp := client.GetSortedFilesWithPagination(ctx, options)
	if errResp != nil {
		log.Printf("Error getting sorted files: %v", errResp.Error)
	} else {
		fmt.Printf("Got %d files (limit: %d, offset: %d)\n", len(files.Items), files.Limit, files.Offset)
		for i, file := range files.Items {
			fmt.Printf("  %d. %s (%s)\n", i+1, file.Name, file.Path)
		}
	}

	fmt.Println()

	// Example 2: Using paginated wrapper with pagination info
	fmt.Println("2. Paginated Wrapper with Pagination Info")
	fmt.Println("------------------------------------------")

	pagedFiles, errResp := client.GetSortedFilesPaged(ctx, options)
	if errResp != nil {
		log.Printf("Error getting paged files: %v", errResp.Error)
	} else {
		fmt.Printf("Files: %d, Pagination Info:\n", len(pagedFiles.Items))
		fmt.Printf("  Limit: %d\n", pagedFiles.Pagination.Limit)
		fmt.Printf("  Offset: %d\n", pagedFiles.Pagination.Offset)
		fmt.Printf("  HasMore: %t\n", pagedFiles.Pagination.HasMore)
		if pagedFiles.Pagination.HasMore {
			fmt.Printf("  NextOffset: %d\n", pagedFiles.Pagination.NextOffset)
		}
	}

	fmt.Println()

	// Example 3: Iterator-based pagination
	fmt.Println("3. Iterator-based Pagination")
	fmt.Println("-----------------------------")

	iteratorOptions := &disk.PaginationOptions{Limit: 3}
	iterator := client.GetSortedFilesIterator(iteratorOptions)
	
	pageNum := 1
	for iterator.HasNext() && pageNum <= 3 { // Limit to 3 pages for demo
		fmt.Printf("Page %d:\n", pageNum)
		
		page, err := iterator.Next(ctx)
		if err != nil {
			log.Printf("Error getting next page: %v", err)
			break
		}

		for i, file := range page.FilesResourceList.Items {
			fmt.Printf("  %d. %s\n", i+1, file.Name)
		}
		
		fmt.Printf("  Pagination: Offset=%d, HasMore=%t\n", 
			page.Pagination.Offset, page.Pagination.HasMore)
		
		pageNum++
		
		// Add a small delay to be respectful to the API
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()

	// Example 4: Different page sizes
	fmt.Println("4. Custom Page Sizes")
	fmt.Println("--------------------")

	pageSizes := []int{2, 10, 50}
	for _, size := range pageSizes {
		opts := &disk.PaginationOptions{Limit: size}
		files, errResp := client.GetSortedFilesWithPagination(ctx, opts)
		if errResp != nil {
			log.Printf("Error with page size %d: %v", size, errResp.Error)
			continue
		}
		fmt.Printf("Page size %d: Got %d files\n", size, len(files.Items))
	}

	fmt.Println()

	// Example 5: Last uploaded resources pagination
	fmt.Println("5. Last Uploaded Resources Pagination")
	fmt.Println("-------------------------------------")

	lastUploadedOptions := &disk.PaginationOptions{Limit: 3}
	lastUploaded, errResp := client.GetLastUploadedResourcesPaged(ctx, lastUploadedOptions)
	if errResp != nil {
		log.Printf("Error getting last uploaded resources: %v", errResp.Error)
	} else {
		fmt.Printf("Last uploaded files: %d\n", len(lastUploaded.Items))
		for i, file := range lastUploaded.Items {
			fmt.Printf("  %d. %s (modified: %s)\n", i+1, file.Name, file.Modified)
		}
		fmt.Printf("HasMore: %t\n", lastUploaded.Pagination.HasMore)
	}

	fmt.Println()

	// Example 6: Public resources pagination
	fmt.Println("6. Public Resources Pagination")
	fmt.Println("------------------------------")

	publicOptions := &disk.PaginationOptions{Limit: 5}
	publicResources, errResp := client.GetPublicResourcesPaged(ctx, publicOptions)
	if errResp != nil {
		log.Printf("Error getting public resources: %v", errResp.Error)
	} else {
		fmt.Printf("Public resources: %d\n", len(publicResources.Items))
		for i, resource := range publicResources.Items {
			fmt.Printf("  %d. %s\n", i+1, resource.Name)
		}
		fmt.Printf("HasMore: %t\n", publicResources.Pagination.HasMore)
	}

	fmt.Println()

	// Example 7: Walking through all pages
	fmt.Println("7. Walking Through All Pages")
	fmt.Println("-----------------------------")

	allFilesIterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 2})
	totalFiles := 0
	pageCount := 0

	for allFilesIterator.HasNext() && pageCount < 5 { // Limit for demo
		page, err := allFilesIterator.Next(ctx)
		if err != nil {
			log.Printf("Error getting page: %v", err)
			break
		}

		pageCount++
		totalFiles += len(page.FilesResourceList.Items)
		
		fmt.Printf("Page %d: %d files (Total so far: %d)\n", 
			pageCount, len(page.FilesResourceList.Items), totalFiles)

		// Add delay to be respectful
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Printf("Processed %d pages with %d total files\n", pageCount, totalFiles)

	fmt.Println()

	// Example 8: Cursor-based pagination (demonstration)
	fmt.Println("8. Cursor-based Pagination Concept")
	fmt.Println("-----------------------------------")

	fmt.Println("Cursor-based pagination is implemented and ready to use")
	fmt.Println("when the Yandex Disk API provides cursor support.")
	fmt.Println("The framework supports both offset/limit and cursor-based pagination.")

	// Create a cursor iterator (won't work with current API but shows the pattern)
	fmt.Println("\nCursor iterator example pattern:")
	fmt.Println("  iterator := client.CreateCursorIterator(limit)")
	fmt.Println("  for iterator.HasNext() {")
	fmt.Println("    page, err := iterator.Next(ctx)")
	fmt.Println("    // process page")
	fmt.Println("  }")

	fmt.Println("\n=== Pagination Examples Complete ===")
}

// Example helper function showing how to collect all results across pages
func collectAllFiles(client *disk.Client, ctx context.Context) ([]*disk.Resource, error) {
	var allFiles []*disk.Resource
	
	iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 20})
	
	for iterator.HasNext() {
		page, err := iterator.Next(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get page: %w", err)
		}
		
		allFiles = append(allFiles, page.FilesResourceList.Items...)
		
		// Add delay to respect API rate limits
		time.Sleep(200 * time.Millisecond)
	}
	
	return allFiles, nil
}

// Example helper function showing how to find specific files with pagination
func findFilesByName(client *disk.Client, ctx context.Context, namePattern string) ([]*disk.Resource, error) {
	var matchingFiles []*disk.Resource
	
	iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 50})
	
	for iterator.HasNext() {
		page, err := iterator.Next(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get page: %w", err)
		}
		
		for _, file := range page.FilesResourceList.Items {
			// Simple name matching - you could use regex or other matching logic
			if len(file.Name) > 0 && file.Name[0:1] == namePattern {
				matchingFiles = append(matchingFiles, file)
			}
		}
		
		time.Sleep(200 * time.Millisecond)
	}
	
	return matchingFiles, nil
}