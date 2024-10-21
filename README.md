# disk

Yandex.Disk API client (WIP)

[REST API Диска](https://yandex.ru/dev/disk/rest/)

<!-- ![GitHub](https://img.shields.io/github/license/ilyabrin/disk) -->
[![Build Status](https://travis-ci.org/ilyabrin/disk.svg?branch=release)](https://travis-ci.org/ilyabrin/disk)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/ilyabrin/disk)
[![Coverage Status](https://coveralls.io/repos/github/ilyabrin/disk/badge.svg?branch=release)](https://coveralls.io/github/ilyabrin/disk?branch=release)

<!-- ![GitHub All Releases](https://img.shields.io/github/downloads/ilyabrin/disk/total) -->
<!-- ![GitHub last commit](https://img.shields.io/github/last-commit/ilyabrin/disk) -->
<!-- ![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/ilyabrin/disk) -->

## Install

```sh
go get -v github.com/ilyabrin/disk
```

## Using

Set the environment variable:

```sh
export YANDEX_DISK_ACCESS_TOKEN=_<your access_token string>_
```

Working example (errors checks omitted):

```go
func main() {

    client := disk.New() // New("or paste your access token here")

    ctx := context.Background()

    disk, err := client.DiskInfo(ctx)
    if err != nil {
        log.Println(err)
    }

    log.Println(disk)
}

```
