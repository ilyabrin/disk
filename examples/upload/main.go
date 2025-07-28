package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ilyabrin/disk"
)

func main() {
	// Check if OAuth token is provided
	token := os.Getenv("YANDEX_DISK_TOKEN")
	if token == "" {
		fmt.Println("Please set YANDEX_DISK_TOKEN environment variable with your OAuth token")
		fmt.Println("You can get one from: https://yandex.ru/dev/disk/poligon/")
		os.Exit(1)
	}

	// Check if file path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run upload_example.go <local-file-path> [remote-path]")
		fmt.Println("Example: go run upload_example.go ./myfile.txt /uploaded/myfile.txt")
		os.Exit(1)
	}

	localPath := os.Args[1]
	remotePath := "/uploaded/" + filepath.Base(localPath)
	if len(os.Args) >= 3 {
		remotePath = os.Args[2]
	}

	// Create Yandex Disk client
	client, err := disk.New(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Set up progress callback
	progressCallback := func(progress disk.UploadProgress) {
		fmt.Printf("\rUploading... %.1f%% (%s / %s)",
			progress.Percentage,
			disk.FormatFileSize(progress.BytesUploaded),
			disk.FormatFileSize(progress.TotalBytes))
	}

	fmt.Printf("Uploading %s to %s...\n", localPath, remotePath)

	// Check file size to determine upload method
	fileSize, err := disk.GetFileSize(localPath)
	if err != nil {
		log.Fatalf("Failed to get file size: %v", err)
	}

	fmt.Printf("File size: %s\n", disk.FormatFileSize(fileSize))

	ctx := context.Background()
	var resource *disk.Resource

	// Use different upload methods based on file size
	if fileSize > 50*1024*1024 { // 50MB
		fmt.Println("Large file detected, using chunked upload...")
		resource, err = client.UploadLargeFileFromPath(ctx, localPath, remotePath, 10, progressCallback)
	} else {
		resource, err = client.UploadFileFromPathWithProgress(ctx, localPath, remotePath, true, progressCallback)
	}

	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	fmt.Printf("\n✅ Upload successful!\n")
	fmt.Printf("Remote path: %s\n", resource.Path)
	fmt.Printf("File name: %s\n", resource.Name)
	fmt.Printf("File size: %d bytes\n", resource.Size)
	fmt.Printf("Created: %s\n", resource.Created)
	fmt.Printf("Modified: %s\n", resource.Modified)

	if resource.PublicURL != "" {
		fmt.Printf("Public URL: %s\n", resource.PublicURL)
	}
}
