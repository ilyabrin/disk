package main

import (
	"context"
	"log"
)

func main() {

	// todo: WithCancel(ctx)
	ctx := context.Background()

	api := New("paste access_token here if not set in YANDEX_DISK_ACCESS_TOKEN envvar")

	diskInfo, err := api.DiskInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(diskInfo)
}
