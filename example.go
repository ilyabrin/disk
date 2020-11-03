package main

import "context"

func main() {

	ctx := context.Background()

	client := New()
	disk, _ := client.DiskInfo(ctx)

	println(string(prettyPrint(disk)))

	link := client.CreateDir(ctx, "000_created_with_api")
	println(string(prettyPrint(link)))

	_ = client.DeleteResource(ctx, "000_created_with_api", false)

}
