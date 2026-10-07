package disk_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ilyabrin/disk"
)

// These examples talk to the real API, so they are compiled but not run.

func Example() {
	// Reads YANDEX_DISK_ACCESS_TOKEN when no token is passed.
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.DiskInfo(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s of %s used\n", disk.FormatFileSize(info.UsedSpace), disk.FormatFileSize(info.TotalSpace))
}

func ExampleClient_GetMetadataWithOptions() {
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}

	// The 20 most recently changed items in Photos.
	folder, errResp := client.GetMetadataWithOptions(context.Background(), "disk:/Photos",
		&disk.ResourceOptions{Limit: 20, Sort: "-modified"})
	if errResp != nil {
		log.Fatalf("%s: %s", errResp.Error, errResp.Description)
	}
	for _, item := range folder.Embedded.Items {
		fmt.Println(item.Type, item.Name, disk.FormatFileSize(item.Size))
	}
}

func ExampleClient_UploadFileFromPath() {
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.UploadFileFromPath(context.Background(), "report.pdf", "disk:/Documents/report.pdf",
		&disk.UploadOptions{Overwrite: true})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("uploaded", res.Path)
}

func ExampleClient_PublishResource() {
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	if _, errResp := client.PublishResource(ctx, "disk:/Documents/report.pdf"); errResp != nil {
		log.Fatal(errResp.Error)
	}
	res, errResp := client.GetMetadata(ctx, "disk:/Documents/report.pdf")
	if errResp != nil {
		log.Fatal(errResp.Error)
	}
	fmt.Println("anyone can open", res.PublicURL)
}

func ExampleClient_GetOperationStatus() {
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	link, errResp := client.CopyResource(ctx, "disk:/Photos", "disk:/Backup/Photos")
	if errResp != nil {
		log.Fatal(errResp.Error)
	}
	// A small folder is copied at once and link points to the copy. A large
	// one is copied in the background and link points to the operation.
	if !strings.Contains(link.Href, "/operations/") {
		fmt.Println("copied")
		return
	}
	op, err := client.GetOperationStatus(ctx, link.Href)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("copy is", op.Status)
}
