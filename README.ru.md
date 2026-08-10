<div align="center">

# disk

**Идиоматичный Go-клиент для [REST API Яндекс Диска](https://yandex.ru/dev/disk-api/doc/ru/).**

[![Go Reference](https://pkg.go.dev/badge/github.com/ilyabrin/disk.svg)](https://pkg.go.dev/github.com/ilyabrin/disk)
[![Run Tests](https://github.com/ilyabrin/disk/actions/workflows/test.yml/badge.svg)](https://github.com/ilyabrin/disk/actions/workflows/test.yml)
[![Security Checks](https://github.com/ilyabrin/disk/actions/workflows/security.yml/badge.svg)](https://github.com/ilyabrin/disk/actions/workflows/security.yml)
[![Coverage Status](https://coveralls.io/repos/github/ilyabrin/disk/badge.svg?branch=release)](https://coveralls.io/github/ilyabrin/disk?branch=release)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilyabrin/disk)](https://goreportcard.com/report/github.com/ilyabrin/disk)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ilyabrin/disk)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

[English version](./README.md) · [Справочник API](https://pkg.go.dev/github.com/ilyabrin/disk) · [Пагинация](./PAGINATION.md)

</div>

---

## Содержание

- [Установка](#установка)
- [Быстрый старт](#быстрый-старт)
- [Аутентификация](#аутентификация)
- [Информация о диске](#информация-о-диске)
- [Файлы и папки](#файлы-и-папки)
- [Загрузка на диск](#загрузка-на-диск)
- [Скачивание](#скачивание)
- [Публичные ресурсы](#публичные-ресурсы)
- [Корзина](#корзина)
- [Пагинация](#пагинация)
- [Пакетные операции](#пакетные-операции)
- [Конфигурация](#конфигурация)
- [Логирование](#логирование)
- [Обработка ошибок](#обработка-ошибок)
- [Асинхронные операции](#асинхронные-операции)
- [Примеры](#примеры)
- [Разработка](#разработка)
- [Лицензия](#лицензия)

---

## Установка

```bash
go get github.com/ilyabrin/disk
```

> [!NOTE]
> Требуется Go 1.23 или новее. Кроме стандартной библиотеки у пакета нет зависимостей времени выполнения.

## Быстрый старт

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ilyabrin/disk"
)

func main() {
	// Если токен не передан явно, он читается из YANDEX_DISK_ACCESS_TOKEN.
	client, err := disk.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	info, err := client.DiskInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Занято %s из %s\n",
		disk.FormatFileSize(int64(info.UsedSpace)),
		disk.FormatFileSize(int64(info.TotalSpace)),
	)
}
```

## Аутентификация

Клиент работает по OAuth-токену. Получить его можно в
[консоли Яндекс OAuth](https://oauth.yandex.ru/) с правами `cloud_api:disk.*`.

```go
client, err := disk.New("y0_AgAAAA...")       // токен явно
client, err := disk.New()                     // из $YANDEX_DISK_ACCESS_TOKEN
```

> [!WARNING]
> Не коммитьте токены. Встроенный логгер скрывает заголовок `Authorization`,
> но за собственное логирование отвечаете вы.

## Информация о диске

```go
info, err := client.DiskInfo(ctx)
if err != nil {
	log.Fatal(err)
}

fmt.Println("Всего:", info.TotalSpace)
fmt.Println("Занято:", info.UsedSpace)
fmt.Println("Корзина:", info.TrashSize)
fmt.Println("Папка «Загрузки»:", info.SystemFolders.Downloads)
```

## Файлы и папки

<details open>
<summary><b>Метаданные</b></summary>

```go
resource, errResp := client.GetMetadata(ctx, "/Документы/отчёт.pdf")
if errResp != nil {
	log.Fatal(errResp.Error)
}
fmt.Println(resource.Name, resource.Size, resource.MimeType)
```

Содержимое папки приходит в поле `_embedded`. Через `GetMetadataWithOptions`
можно управлять пагинацией, сортировкой и составом ответа:

```go
folder, errResp := client.GetMetadataWithOptions(ctx, "/Фото", &disk.ResourceOptions{
	Limit:       100,
	Offset:      0,
	Sort:        "-modified",                       // «-» разворачивает порядок
	PreviewSize: "M",
	PreviewCrop: true,
	Fields:      []string{"name", "_embedded.items.name", "_embedded.items.size"},
})

for _, item := range folder.Embedded.Items {
	fmt.Println(item.Type, item.Name)
}
```

</details>

<details>
<summary><b>Создание, копирование, перемещение, удаление</b></summary>

```go
// Создать папку
link, errResp := client.CreateDir(ctx, "/Отчёты")

// Копировать
link, errResp = client.CopyResource(ctx, "/a.txt", "/backup/a.txt")

// Переместить или переименовать
link, errResp = client.MoveResource(ctx, "/a.txt", "/archive/a-2026.txt")

// Удалить в корзину
err := client.DeleteResource(ctx, "/a.txt", false)

// Удалить безвозвратно
err = client.DeleteResource(ctx, "/a.txt", true)
```

</details>

<details>
<summary><b>Пользовательские атрибуты</b></summary>

```go
resource, errResp := client.UpdateMetadata(ctx, "/отчёт.pdf",
	map[string]map[string]string{
		"custom_properties": {
			"project": "apollo",
			"status":  "final",
		},
	})
```

</details>

<details>
<summary><b>Плоский список файлов и последние загрузки</b></summary>

```go
// Все файлы, сначала новые, только изображения и видео
files, errResp := client.GetSortedFilesWithOptions(ctx,
	&disk.PaginationOptions{Limit: 50},
	&disk.FilesOptions{
		MediaType: []string{"image", "video"},
		Sort:      "-created",
	})

// Последние загруженные ресурсы
recent, errResp := client.GetLastUploadedResources(ctx)
```

</details>

## Загрузка на диск

| Метод | Когда использовать |
| --- | --- |
| `UploadFileFromPath` | Любой локальный файл, полный контроль над опциями |
| `UploadFileFromPathWithProgress` | Небольшие и средние файлы с прогрессом |
| `UploadLargeFileFromPath` | Большие файлы, прогресс по чанкам |
| `UploadFile` | Загрузка по URL силами самого Яндекса |

```go
resource, err := client.UploadFileFromPath(ctx, "./отчёт.pdf", "/Документы/отчёт.pdf",
	&disk.UploadOptions{Overwrite: true})
```

С прогрессом:

```go
resource, err := client.UploadFileFromPathWithProgress(ctx,
	"./video.mp4", "/Видео/video.mp4", true,
	func(p disk.UploadProgress) {
		fmt.Printf("\r%.1f%% (%s / %s)",
			p.Percentage,
			disk.FormatFileSize(p.BytesUploaded),
			disk.FormatFileSize(p.TotalBytes),
		)
	})
```

Большие файлы, отчёт раз в 10 МБ:

```go
resource, err := client.UploadLargeFileFromPath(ctx,
	"./archive.zip", "/Бэкапы/archive.zip", 10,
	func(p disk.UploadProgress) {
		log.Printf("загружено %s", disk.FormatFileSize(p.BytesUploaded))
	})
```

Чтобы Яндекс сам скачал файл по ссылке, не пропуская трафик через ваш процесс:

```go
link, errResp := client.UploadFile(ctx, "/Загрузки/image.jpg", "https://example.com/image.jpg")
```

> [!TIP]
> `UploadFile` работает асинхронно — узнать, что файл долетел, можно опросив
> возвращённую ссылку через [`GetOperationStatus`](#асинхронные-операции).

## Скачивание

```go
err := client.DownloadFileToPath(ctx, "/Фото/image.jpg", "./image.jpg",
	&disk.DownloadOptions{Overwrite: true})
```

С прогрессом:

```go
err := client.DownloadFileToPathWithProgress(ctx,
	"/Видео/video.mp4", "./video.mp4", true,
	func(p disk.DownloadProgress) {
		if p.TotalBytes > 0 {
			fmt.Printf("\r%.1f%%", p.Percentage)
		}
	})
```

Нужна сама ссылка (для CDN, редиректа в браузере или собственной качалки)?

```go
link, errResp := client.GetDownloadURL(ctx, "/Фото/image.jpg")
fmt.Println(link.Href) // короткоживущая, одноразовая
```

## Публичные ресурсы

```go
// Опубликовать и снять публикацию
link, errResp := client.PublishResource(ctx, "/Фото/image.jpg")
link, errResp = client.UnpublishResource(ctx, "/Фото/image.jpg")

// Всё, что вы опубликовали
list, errResp := client.GetPublicResources(ctx)
```

Чтение чужого опубликованного ресурса по ключу или ссылке:

```go
resource, errResp := client.GetMetadataForPublicResource(ctx, "https://yadi.sk/d/abc123")

// Заглянуть внутрь опубликованной папки
resource, errResp = client.GetMetadataForPublicResourceWithOptions(ctx, "https://yadi.sk/d/abc123",
	&disk.PublicResourceOptions{
		Path:  "/subfolder",
		Sort:  "name",
		Limit: 50,
	})

// Скачать конкретный файл из опубликованной папки
link, errResp := client.GetDownloadURLForPublicResourceAt(ctx, "https://yadi.sk/d/abc123", "/subfolder/file.txt")

// Сохранить его к себе в «Загрузки» под новым именем
link, errResp = client.SavePublicResourceWithOptions(ctx, "https://yadi.sk/d/abc123",
	&disk.SavePublicResourceOptions{Path: "/subfolder/file.txt", Name: "copy.txt"})
```

## Корзина

```go
// Просмотр
trash, err := client.ListTrashResources(ctx, "", 100, 0)

// Метаданные одного элемента
item, err := client.GetTrashResourceMetadata(ctx, "отчёт.pdf", nil)

// Восстановить, при желании переименовав
link, err := client.RestoreFromTrash(ctx, "отчёт.pdf", false, "отчёт-восстановлен.pdf")

// Очистить один путь или всю корзину, если передать ""
err = client.EmptyTrash(ctx, "", false)
```

> [!NOTE]
> `forceAsync: true` в `EmptyTrash` заставляет API всегда отвечать `202` со ссылкой
> на операцию, вместо того чтобы блокироваться на большом удалении.

## Пагинация

Доступны три подхода — от низкоуровневого к высокоуровневому. Подробности в [PAGINATION.md](./PAGINATION.md).

<details open>
<summary><b>1. Явные limit/offset</b></summary>

```go
files, errResp := client.GetSortedFilesWithPagination(ctx, &disk.PaginationOptions{
	Limit:  50,
	Offset: 100,
})
```

</details>

<details>
<summary><b>2. Страница с метаданными</b></summary>

```go
page, errResp := client.GetSortedFilesPaged(ctx, &disk.PaginationOptions{Limit: 50})
fmt.Println(page.Pagination.HasMore, page.Pagination.NextOffset)
```

</details>

<details>
<summary><b>3. Итератор</b></summary>

```go
it := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 100})

for it.HasNext() {
	page, err := it.Next(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range page.Items {
		fmt.Println(file.Name)
	}
}
```

</details>

## Пакетные операции

Пакетные методы выполняют операции параллельно и собирают результат по каждому
элементу, а не падают на первой же ошибке.

```go
status, err := client.BatchDeleteFiles(ctx,
	[]string{"/tmp/a.txt", "/tmp/b.txt", "/tmp/c.txt"},
	&disk.BatchDeleteOptions{
		BatchOptions: disk.BatchOptions{
			MaxConcurrency:  4,
			ContinueOnError: true,
			Progress: func(s disk.BatchOperationStatus) {
				fmt.Printf("\r%d/%d", s.Completed, s.Total)
			},
		},
		Permanently: false,
	})

fmt.Println(status.GetSummary())

for _, failure := range status.GetFailedOperations() {
	log.Printf("%s: %v", failure.Path, failure.Error)
}

// Повторить только то, что упало
status, err = client.RetryFailedOperations(ctx, status, 2)
```

Доступны `BatchDeleteFiles`, `BatchCopyFiles`, `BatchMoveFiles`,
`BatchUpdateMetadata`, а также обёртки `BatchRenameFiles`,
`BatchMoveToDirectory`, `BatchCopyToDirectory` и варианты `*Simple`.

## Конфигурация

```go
client, err := disk.NewWithConfig(&disk.ClientConfig{
	DefaultTimeout:     60 * time.Second,
	MaxRetries:         3,
	RetryBackoff:       200 * time.Millisecond,
	EnableDebugLogging: true,
	Logger:             disk.DefaultLoggerConfig(),
}, "ваш-токен")
```

| Поле | По умолчанию | Значение |
| --- | --- | --- |
| `DefaultTimeout` | `30s` | Таймаут запроса, если у контекста нет дедлайна |
| `MaxRetries` | `3` | Дополнительные попытки для повторяемых запросов |
| `RetryBackoff` | `200ms` | Базовая пауза между попытками, удваивается |
| `EnableDebugLogging` | `false` | Переводит логгер в `DEBUG` и подробный режим |
| `Logger` | см. ниже | Настройки логгера |
| `BaseURL` | API Яндекса | Другой адрес API (тесты, прокси) |

> [!IMPORTANT]
> Повторяются только запросы **без тела** — `GET`, `DELETE` и те `PUT`/`POST`,
> что передают параметры в query string. Тело-`io.Reader` нельзя перемотать,
> поэтому такой запрос отправляется ровно один раз. Повтор происходит на ошибках
> соединения, `429` и `5xx`.

Таймаут можно задать и на конкретный вызов через контекст:

```go
ctx, cancel := disk.WithTimeout(10 * time.Second)
defer cancel()

info, err := client.DiskInfo(ctx)
```

## Логирование

```go
client.SetLogLevel(disk.DEBUG)   // DEBUG, INFO, WARN, ERROR, SILENT
client.SetVerbose(true)          // подробности запросов и ответов
client.SetLogOutput(os.Stderr)   // любой io.Writer
```

Чувствительные значения заголовков (`Authorization`, токены) скрываются до
попадания в лог.

## Обработка ошибок

В библиотеке два соглашения об ошибках, и какое сработает — зависит от вызова:

| Тип | Где | Как обрабатывать |
| --- | --- | --- |
| `*ErrorResponse` | Ресурсы, публичные ресурсы, пагинация | Не-`nil` означает ошибку; смотрите `.Error` и `.Description` |
| `error` | Информация о диске, загрузка, скачивание, корзина, пакеты | Обычный Go-подход, ошибки обёрнуты через `%w` |

```go
resource, errResp := client.GetMetadata(ctx, "/нет-такого.txt")
if errResp != nil {
	log.Printf("%s: %s", errResp.Error, errResp.Description)
	return
}

if err := client.DownloadFileToPath(ctx, "/a.txt", "./a.txt", nil); err != nil {
	log.Fatal(err)
}
```

## Асинхронные операции

Копирование, перемещение, сохранение публичного ресурса и очистка корзины могут
ответить `202 Accepted` со ссылкой на фоновую операцию. Опрашивайте её, пока
статус не перестанет быть `in-progress`:

```go
link, errResp := client.CopyResource(ctx, "/большая-папка", "/backup/большая-папка")
if errResp != nil {
	log.Fatal(errResp.Error)
}

for {
	// Принимает и идентификатор операции, и полный href из ответа.
	operation, err := client.GetOperationStatus(ctx, link.Href)
	if err != nil {
		log.Fatal(err)
	}
	if operation.Status != disk.OperationInProgress {
		fmt.Println("готово:", operation.Status)
		break
	}
	time.Sleep(time.Second)
}
```

Для пакетных вызовов опрос берёт на себя `WaitForBatchOperation`:

```go
err := client.WaitForBatchOperation(ctx, status, time.Second)
```

## Примеры

Готовые программы лежат в [examples/](./examples):

| Пример | Что показывает |
| --- | --- |
| [demo](./examples/demo) | Информация о диске, метаданные, операции с файлами и папками |
| [upload](./examples/upload) | Загрузка с отображением прогресса |
| [pagination](./examples/pagination) | Все три способа пагинации |

```bash
export YANDEX_DISK_ACCESS_TOKEN=ваш-токен
go run ./examples/demo
```

## Разработка

```bash
go test ./...                                  # прогнать тесты
go test -race -cover ./...                     # с детектором гонок
go vet ./...                                   # статические проверки
golangci-lint run                              # полный линт (см. .golangci.yml)
```

CI на каждый push и pull request запускает тесты, `go vet`, CodeQL, `govulncheck`
и gosec; находки gosec публикуются как code scanning alerts.

Пул-реквесты и issue приветствуются.

## Лицензия

[MIT](./LICENSE) © Ilya Brin
