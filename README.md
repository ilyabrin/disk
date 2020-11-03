# disk
Yandex.Disk API client

## Install
```sh
go get -v github.com/ilyabrin/disk
```

## Using

Set the environment variable:
```sh
> export YANDEX_DISK_ACCESS_TOKEN=_<your access_token string>_
```

Working example (errors checks omitted):
```go
package main

import (
    "context"
    disk "github.com/ilyabrin/disk"
)

func main() {
	ctx := context.Background()

    client := disk.New()

	disk, _ := client.DiskInfo(ctx)
    link, _ := client.CreateDir(ctx, "000_created_with_api")

	_ = client.DeleteResource(ctx, "000_created_with_api", false)
}
```

_WIP_