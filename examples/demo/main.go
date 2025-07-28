package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyabrin/disk"
)

func main() {
	fmt.Println("🚀 Yandex Disk Upload Demo")
	fmt.Println("==========================")

	// Demo of utility functions that don't require a token
	fmt.Println("\n📁 File Utilities Demo:")

	// Test file size detection
	if len(os.Args) > 1 {
		filePath := os.Args[1]

		// Validate file path for Yandex Disk
		if err := disk.ValidateFilePath("/uploaded/" + filepath.Base(filePath)); err != nil {
			fmt.Printf("❌ Path validation failed: %v\n", err)
		} else {
			fmt.Printf("✅ Path is valid for Yandex Disk\n")
		}

		// Get file size
		if size, err := disk.GetFileSize(filePath); err != nil {
			fmt.Printf("❌ Could not get file size: %v\n", err)
		} else {
			fmt.Printf("📊 File size: %s (%d bytes)\n", disk.FormatFileSize(size), size)

			// Recommend upload method based on size
			if size > 50*1024*1024 {
				fmt.Printf("💡 Recommendation: Use UploadLargeFileFromPath() for files > 50MB\n")
			} else {
				fmt.Printf("💡 Recommendation: Use UploadFileFromPath() for smaller files\n")
			}
		}

		// Note: MIME type detection requires a client instance
		// For demo purposes, we'll skip this since we don't have a token
		fmt.Printf("🎭 MIME type detection available via client.DetectMimeType()\n")
	} else {
		fmt.Println("Usage: go run demo.go <file-path>")
		fmt.Println("Example: go run demo.go test_file.txt")
		fmt.Println("\nThis demo shows file utilities that work without a Yandex Disk token.")
	}

	fmt.Println("\n📚 File Size Formatting Examples:")
	sizes := []int64{512, 1024, 1536, 1024 * 1024, 5 * 1024 * 1024, 1024 * 1024 * 1024}
	for _, size := range sizes {
		fmt.Printf("  %d bytes → %s\n", size, disk.FormatFileSize(size))
	}

	fmt.Println("\n🔑 For actual uploads, you need:")
	fmt.Println("  1. Set YANDEX_DISK_TOKEN environment variable")
	fmt.Println("  2. Get token from: https://yandex.ru/dev/disk/poligon/")
	fmt.Println("  3. Use upload_example.go for real uploads")

	fmt.Println("\n✨ Available Upload Methods:")
	fmt.Println("  • client.UploadFileFromPath() - Basic upload with options")
	fmt.Println("  • client.UploadFileFromPathWithProgress() - Upload with progress callback")
	fmt.Println("  • client.UploadLargeFileFromPath() - Chunked upload for large files")
}
