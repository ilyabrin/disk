# Клиентская библиотека Яндекс.Диска на Go

[🇬🇧 English Version](./README.md)

[![Версия Go](https://img.shields.io/badge/go-%3E%3D1.20-blue.svg)](https://golang.org/)
[![Лицензия](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)

Комплексная, готовая к промышленной эксплуатации клиентская библиотека на Go для [REST API Яндекс.Диска](https://yandex.ru/dev/disk/rest/). Эта библиотека предоставляет чистый, идиоматичный Go-интерфейс для взаимодействия с облачным хранилищем Яндекс.Диск.

## 🌟 Возможности

### Основной функционал

- ✅ **Полное покрытие API** - Полная поддержка всех эндпоинтов REST API Яндекс.Диска
- 🔐 **OAuth2 аутентификация** - Простая аутентификация на основе токенов
- 📁 **Операции с файлами** - Загрузка, скачивание, копирование, перемещение и удаление файлов
- 📊 **Управление метаданными** - Получение и обновление метаданных файлов/папок
- 🗑️ **Управление корзиной** - Перемещение в корзину, восстановление и окончательное удаление
- 🌐 **Публичные ссылки** - Создание и управление публичными ссылками на файлы и папки
- 📦 **Информация о диске** - Получение информации о месте на диске, квоте и системных папках

### Расширенные возможности

- 🔄 **Поддержка пагинации** - Множество стратегий пагинации (на основе смещения и паттерн итератора)
- 📦 **Пакетные операции** - Эффективная обработка нескольких файлов с параллельным выполнением
- 📤 **Умная загрузка** - Автоматическая обработка больших файлов с отслеживанием прогресса
- ⚡ **Поддержка контекста** - Полная интеграция с context.Context для таймаутов и отмены операций
- 📝 **Всеобъемлющее логирование** - Встроенное структурированное логирование с несколькими уровнями важности
- 🔒 **Безопасность** - Валидация и санитизация путей для предотвращения атак
- ⚙️ **Настраиваемость** - Обширные параметры конфигурации для таймаутов, повторов и многого другого
- 🧪 **Хорошо протестирована** - Высокое покрытие тестами с всесторонними модульными тестами

## 📋 Содержание

- [Установка](#-установка)
- [Быстрый старт](#-быстрый-старт)
- [Аутентификация](#-аутентификация)
- [Базовое использование](#-базовое-использование)
  - [Информация о диске](#информация-о-диске)
  - [Операции с файлами](#операции-с-файлами)
  - [Операции с папками](#операции-с-папками)
  - [Загрузка файлов](#загрузка-файлов)
  - [Скачивание файлов](#скачивание-файлов)
  - [Публичные ресурсы](#публичные-ресурсы)
  - [Управление корзиной](#управление-корзиной)
- [Расширенные возможности](#-расширенные-возможности)
  - [Пагинация](#пагинация)
  - [Пакетные операции](#пакетные-операции)
  - [Отслеживание прогресса](#отслеживание-прогресса)
- [Пользовательская конфигурация](#пользовательская-конфигурация)
  - [Логирование](#логирование)
- [Обработка ошибок](#️-обработка-ошибок)
- [Примеры](#-примеры)
- [Справочник API](#-справочник-api)
- [Тестирование](#-тестирование)
- [Вклад в проект](#-вклад-в-проект)
- [Лицензия](#-лицензия)

## 📦 Установка

```bash
go get github.com/ilyabrin/disk
```

**Требования:**

- Go 1.20 или выше

## 🚀 Быстрый старт

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ilyabrin/disk"
)

func main() {
    // Создание нового клиента с вашим OAuth-токеном
    client := disk.NewClient("ВАШ_OAUTH_ТОКЕН")

    // Создание контекста с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Получение информации о диске
    diskInfo, err := client.GetDisk(ctx)
    if err != nil {
        log.Fatalf("Не удалось получить информацию о диске: %v", err)
    }

    fmt.Printf("Общее место: %d байт\n", diskInfo.TotalSpace)
    fmt.Printf("Использовано: %d байт\n", diskInfo.UsedSpace)
    fmt.Printf("Доступно: %d байт\n", diskInfo.TotalSpace-diskInfo.UsedSpace)
}
```

## 🔐 Аутентификация

Для использования этой библиотеки вам нужен OAuth-токен от Яндекса. Вот как его получить:

1. Перейдите на [Яндекс OAuth](https://oauth.yandex.ru/)
2. Зарегистрируйте ваше приложение
3. Запросите доступ к скоупам `cloud_api:disk.read` и `cloud_api:disk.write`
4. Получите ваш OAuth-токен

```go
// Создание клиента с вашим токеном
client := disk.NewClient("ВАШ_OAUTH_ТОКЕН")

// Или с пользовательской конфигурацией
config := disk.DefaultClientConfig()
config.DefaultTimeout = 60 * time.Second
client = disk.NewClientWithConfig("ВАШ_OAUTH_ТОКЕН", config)
```

## 💡 Базовое использование

### Информация о диске

```go
// Получение информации о диске
diskInfo, err := client.GetDisk(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Общее место: %d\n", diskInfo.TotalSpace)
fmt.Printf("Использовано: %d\n", diskInfo.UsedSpace)
fmt.Printf("Размер корзины: %d\n", diskInfo.TrashSize)
fmt.Printf("Платная подписка: %t\n", diskInfo.IsPaid)
```

### Операции с файлами

#### Получение метаданных файла

```go
// Получение метаданных файла или папки
resource, errResp := client.GetMetadata(ctx, "/путь/к/файлу.txt")
if errResp != nil {
    log.Fatalf("Ошибка: %s", errResp.Error)
}

fmt.Printf("Имя: %s\n", resource.Name)
fmt.Printf("Размер: %d байт\n", resource.Size)
fmt.Printf("Тип: %s\n", resource.Type)
fmt.Printf("Изменён: %s\n", resource.Modified)
```

#### Копирование файла

```go
// Копирование файла или папки
link, err := client.CopyResource(ctx, "/источник/файл.txt", "/назначение/файл.txt", false)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Ссылка на операцию копирования: %s\n", link.Href)
```

#### Перемещение/переименование файла

```go
// Перемещение или переименование файла
link, err := client.MoveResource(ctx, "/старый/путь/файл.txt", "/новый/путь/файл.txt", false)
if err != nil {
    log.Fatal(err)
}
```

#### Удаление файла

```go
// Удаление файла (перемещение в корзину)
err := client.DeleteResource(ctx, "/путь/к/файлу.txt", false)
if err != nil {
    log.Fatal(err)
}

// Окончательное удаление файла
err = client.DeleteResource(ctx, "/путь/к/файлу.txt", true)
```

### Операции с папками

#### Создание папки

```go
// Создание новой папки
link, err := client.CreateFolder(ctx, "/путь/к/новой/папке")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Папка создана: %s\n", link.Href)
```

#### Список содержимого папки

```go
// Получение содержимого папки
resource, errResp := client.GetMetadata(ctx, "/путь/к/папке")
if errResp != nil {
    log.Fatal(errResp.Error)
}

// Перебор элементов
if resource.Embedded != nil {
    for _, item := range resource.Embedded.Items {
        fmt.Printf("- %s (%s)\n", item.Name, item.Type)
    }
}
```

### Загрузка файлов

#### Простая загрузка

```go
// Загрузка небольшого файла
options := &disk.UploadOptions{
    Overwrite: true,
    Progress: func(progress disk.UploadProgress) {
        fmt.Printf("Загружено: %.2f%%\n", progress.Percentage)
    },
}

resource, err := client.UploadFileFromPath(ctx, "локальный/файл.txt", "/диск/файл.txt", options)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Файл загружен: %s\n", resource.Name)
```

#### Загрузка большого файла

```go
// Загрузка большого файла с автоматическим разбиением на части
resource, err := client.UploadLargeFileFromPath(ctx, "большой-файл.zip", "/диск/большой-файл.zip", nil)
if err != nil {
    log.Fatal(err)
}
```

#### Загрузка из Reader

```go
// Загрузка из io.Reader
file, _ := os.Open("файл.txt")
defer file.Close()

resource, err := client.UploadFile(ctx, file, "/диск/файл.txt", options)
```

### Скачивание файлов

```go
// Скачивание файла
err := client.DownloadFile(ctx, "/диск/файл.txt", "локальный/файл.txt")
if err != nil {
    log.Fatal(err)
}

// Или получение ссылки на скачивание
link, errResp := client.GetDownloadURL(ctx, "/диск/файл.txt")
if errResp != nil {
    log.Fatal(errResp.Error)
}
fmt.Printf("Ссылка на скачивание: %s\n", link.Href)
```

### Публичные ресурсы

#### Публикация ресурса

```go
// Сделать файл или папку публичными
link, err := client.PublishResource(ctx, "/путь/к/файлу.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Публичная ссылка: %s\n", link.Href)
```

#### Снятие публикации ресурса

```go
// Удаление публичного доступа
link, err := client.UnpublishResource(ctx, "/путь/к/файлу.txt")
if err != nil {
    log.Fatal(err)
}
```

#### Доступ к публичному ресурсу

```go
// Получение метаданных публичного ресурса
resource, errResp := client.GetMetadataForPublicResource(ctx, "публичный-ключ")
if errResp != nil {
    log.Fatal(errResp.Error)
}

// Скачивание публичного ресурса
link, errResp := client.GetDownloadURLForPublicResource(ctx, "публичный-ключ")
```

#### Список публичных ресурсов

```go
// Получение всех публичных ресурсов
options := &disk.PaginationOptions{Limit: 20}
pagedResources, errResp := client.GetPublicResourcesPaged(ctx, options)
if errResp != nil {
    log.Fatal(errResp.Error)
}

for _, resource := range pagedResources.Items {
    fmt.Printf("Публичный: %s (%s)\n", resource.Name, resource.PublicURL)
}
```

### Управление корзиной

#### Перемещение в корзину

```go
// Удаление файла (перемещение в корзину)
err := client.DeleteResource(ctx, "/путь/к/файлу.txt", false)
```

#### Список содержимого корзины

```go
// Список элементов в корзине
trashList, err := client.ListTrashResources(ctx, "", 20, 0)
if err != nil {
    log.Fatal(err)
}

for _, item := range trashList.Embedded.Items {
    fmt.Printf("Корзина: %s (удалён: %s)\n", item.Name, item.Deleted)
}
```

#### Восстановление из корзины

```go
// Восстановление файла из корзины
link, err := client.RestoreFromTrash(ctx, "/путь/к/файлу.txt", false, "")
if err != nil {
    log.Fatal(err)
}
```

#### Очистка корзины

```go
// Окончательное удаление всей корзины
link, err := client.EmptyTrash(ctx)
if err != nil {
    log.Fatal(err)
}
```

#### Окончательное удаление из корзины

```go
// Окончательное удаление конкретного элемента из корзины
link, err := client.DeleteFromTrash(ctx, "/путь/к/файлу.txt")
```

## 🔥 Расширенные возможности

### Пагинация

Библиотека предоставляет всестороннюю поддержку пагинации с множеством стратегий. Подробную документацию смотрите в [PAGINATION.md](./PAGINATION.md).

#### Базовая пагинация

```go
// Получение файлов с пагинацией
options := &disk.PaginationOptions{
    Limit:  20,
    Offset: 0,
}
files, errResp := client.GetSortedFilesWithPagination(ctx, options)
```

#### Расширенная пагинация с метаданными

```go
// Получение информации о пагинации
pagedFiles, errResp := client.GetSortedFilesPaged(ctx, options)
if errResp == nil {
    fmt.Printf("Всего элементов: %d\n", len(pagedFiles.Items))
    fmt.Printf("Есть ещё: %t\n", pagedFiles.Pagination.HasMore)
    if pagedFiles.Pagination.HasMore {
        fmt.Printf("Следующее смещение: %d\n", pagedFiles.Pagination.NextOffset)
    }
}
```

#### Паттерн итератора

```go
// Использование итератора для автоматической пагинации
iterator := client.GetSortedFilesIterator(&disk.PaginationOptions{Limit: 50})

for iterator.HasNext() {
    page, err := iterator.Next(ctx)
    if err != nil {
        log.Printf("Ошибка: %v", err)
        break
    }
    
    for _, file := range page.FilesResourceList.Items {
        fmt.Printf("Файл: %s (%d байт)\n", file.Name, file.Size)
    }
    
    // Ограничение частоты запросов
    time.Sleep(200 * time.Millisecond)
}
```

### Пакетные операции

Эффективная обработка нескольких файлов с параллельным выполнением.

#### Пакетное удаление

```go
paths := []string{
    "/файл1.txt",
    "/файл2.txt",
    "/папка/файл3.txt",
}

options := &disk.BatchDeleteOptions{
    BatchOptions: disk.BatchOptions{
        MaxConcurrency:  5,
        ContinueOnError: true,
        Progress: func(status disk.BatchOperationStatus) {
            fmt.Printf("Прогресс: %d/%d (%.1f%%)\n", 
                status.Completed, status.Total, status.Percentage)
        },
    },
    Permanently: false,
}

status, err := client.BatchDeleteFiles(ctx, paths, options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Успешно: %d, Неудачно: %d\n", status.Successful, status.Failed)
```

#### Пакетное копирование

```go
sourceDestMap := map[string]string{
    "/источник/файл1.txt": "/резервная/файл1.txt",
    "/источник/файл2.txt": "/резервная/файл2.txt",
}

options := &disk.BatchCopyMoveOptions{
    BatchOptions: disk.BatchOptions{
        MaxConcurrency: 3,
    },
    Overwrite: false,
}

status, err := client.BatchCopyFiles(ctx, sourceDestMap, options)
```

#### Пакетное перемещение

```go
status, err := client.BatchMoveFiles(ctx, sourceDestMap, options)
```

### Отслеживание прогресса

Отслеживание прогресса загрузки/скачивания в реальном времени.

```go
options := &disk.UploadOptions{
    Progress: func(progress disk.UploadProgress) {
        percentage := progress.Percentage
        bytes := progress.BytesUploaded
        total := progress.TotalBytes
        
        fmt.Printf("\rЗагрузка: %.2f%% (%d/%d байт)", 
            percentage, bytes, total)
    },
}

resource, err := client.UploadFileFromPath(ctx, localPath, remotePath, options)
```

### Пользовательская конфигурация

```go
// Создание пользовательской конфигурации
config := &disk.ClientConfig{
    DefaultTimeout:     60 * time.Second,
    MaxRetries:         3,
    EnableDebugLogging: true,
    Logger: &disk.LoggerConfig{
        Enabled:      true,
        Level:        disk.LogLevelInfo,
        IncludeTime:  true,
        ColorEnabled: true,
    },
}

client := disk.NewClientWithConfig("ВАШ_ТОКЕН", config)
```

### Логирование

Библиотека включает всесторонние возможности логирования.

```go
// Доступ к логгеру
client.Logger.Info("Операция запущена")
client.Logger.Debug("Отладочная информация: %v", data)
client.Logger.Error("Произошла ошибка: %v", err)

// Изменение уровня логирования
client.Logger.SetLevel(disk.LogLevelDebug)

// Включение/отключение логирования
client.Logger.SetEnabled(true)

// Включение цветного вывода
client.Logger.SetColorEnabled(true)
```

**Доступные уровни логирования:**

- `LogLevelDebug` - Подробная отладочная информация
- `LogLevelInfo` - Общие информационные сообщения
- `LogLevelWarn` - Предупреждающие сообщения
- `LogLevelError` - Сообщения об ошибках

## ⚠️ Обработка ошибок

Библиотека предоставляет подробную информацию об ошибках через структурированные ответы об ошибках.

```go
resource, errResp := client.GetMetadata(ctx, "/путь/к/файлу")
if errResp != nil {
    fmt.Printf("Ошибка: %s\n", errResp.Error)
    fmt.Printf("Описание: %s\n", errResp.Description)
    fmt.Printf("Сообщение: %s\n", errResp.Message)
    
    // Обработка конкретных ошибок
    switch errResp.Error {
    case "DiskNotFoundError":
        fmt.Println("Файл или папка не найдены")
    case "UnauthorizedError":
        fmt.Println("Недействительный или просроченный токен")
    case "DiskPathPointsToExistentDirectoryError":
        fmt.Println("Путь уже существует")
    default:
        fmt.Printf("Неизвестная ошибка: %s\n", errResp.Error)
    }
    return
}
```

**Распространённые типы ошибок:**

- `UnauthorizedError` - Недействительный или просроченный OAuth-токен
- `DiskNotFoundError` - Ресурс не найден
- `DiskPathPointsToExistentDirectoryError` - Путь уже существует
- `FieldValidationError` - Недопустимые входные параметры
- `LockedError` - Ресурс заблокирован
- `LimitExceededError` - Превышен лимит запросов

## 📚 Примеры

Полные рабочие примеры доступны в директории `examples/`:

- **[demo/main.go](./examples/demo/main.go)** - Демо утилит для файлов и валидации
- **[pagination/main.go](./examples/pagination/main.go)** - Паттерны пагинации и итераторы
- **[upload/main.go](./examples/upload/main.go)** - Загрузка файлов с отслеживанием прогресса

Запуск примера:

```bash
cd examples/upload
go run main.go
```

## 📖 Справочник API

### Методы клиента

#### Информация о диске

- `GetDisk(ctx) (*Disk, error)` - Получить информацию о диске

#### Операции с файлами и папками

- `GetMetadata(ctx, path) (*Resource, *ErrorResponse)` - Получить метаданные ресурса
- `CreateFolder(ctx, path) (*Link, error)` - Создать папку
- `CopyResource(ctx, from, to, overwrite) (*Link, error)` - Копировать ресурс
- `MoveResource(ctx, from, to, overwrite) (*Link, error)` - Переместить ресурс
- `DeleteResource(ctx, path, permanently) error` - Удалить ресурс

#### Загрузка и скачивание

- `UploadFileFromPath(ctx, local, remote, options) (*Resource, error)` - Загрузить файл
- `UploadLargeFileFromPath(ctx, local, remote, options) (*Resource, error)` - Загрузить большой файл
- `UploadFile(ctx, reader, remote, options) (*Resource, error)` - Загрузить из reader
- `DownloadFile(ctx, remote, local) error` - Скачать файл
- `GetDownloadURL(ctx, path) (*Link, *ErrorResponse)` - Получить ссылку на скачивание

#### Публичные ресурсы

- `PublishResource(ctx, path) (*Link, error)` - Опубликовать ресурс
- `UnpublishResource(ctx, path) (*Link, error)` - Снять публикацию ресурса
- `GetMetadataForPublicResource(ctx, publicKey) (*PublicResource, *ErrorResponse)`
- `GetDownloadURLForPublicResource(ctx, publicKey) (*Link, *ErrorResponse)`
- `GetPublicResources(ctx) (*PublicResourcesList, *ErrorResponse)`
- `GetPublicResourcesPaged(ctx, options) (*PagedPublicResourcesList, *ErrorResponse)`
- `GetPublicResourcesIterator(options) *OffsetPaginationIterator[*PublicResourcesList]`

#### Операции с корзиной

- `ListTrashResources(ctx, path, limit, offset) (*TrashResourceList, error)`
- `RestoreFromTrash(ctx, path, overwrite, name) (*Link, error)`
- `DeleteFromTrash(ctx, path) (*Link, error)`
- `EmptyTrash(ctx) (*Link, error)`

#### Пакетные операции

- `BatchDeleteFiles(ctx, paths, options) (*BatchOperationStatus, error)`
- `BatchCopyFiles(ctx, sourceDestMap, options) (*BatchOperationStatus, error)`
- `BatchMoveFiles(ctx, sourceDestMap, options) (*BatchOperationStatus, error)`

#### Пагинация

- `GetSortedFiles(ctx) (*FilesResourceList, *ErrorResponse)`
- `GetSortedFilesWithPagination(ctx, options) (*FilesResourceList, *ErrorResponse)`
- `GetSortedFilesPaged(ctx, options) (*PagedFilesResourceList, *ErrorResponse)`
- `GetSortedFilesIterator(options) *OffsetPaginationIterator[*FilesResourceList]`
- `GetLastUploadedResources(ctx) (*LastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesWithPagination(ctx, options) (*LastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesPaged(ctx, options) (*PagedLastUploadedResourceList, *ErrorResponse)`
- `GetLastUploadedResourcesIterator(options) *OffsetPaginationIterator[*LastUploadedResourceList]`

#### Операции

- `OperationStatus(ctx, operationID) (any, *http.Response, error)` - Проверить статус асинхронной операции

Для подробной документации по API пагинации смотрите [PAGINATION.md](./PAGINATION.md).

## 🧪 Тестирование

Библиотека включает всесторонние модульные тесты.

```bash
# Запуск всех тестов
go test -v

# Запуск с покрытием
go test -v -cover

# Создание отчёта о покрытии
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Запуск конкретных тестов
go test -v -run TestPagination
go test -v -run TestBatch
```

## 🤝 Вклад в проект

Вклады приветствуются! Пожалуйста, не стесняйтесь отправлять Pull Request. Для больших изменений, пожалуйста, сначала откройте issue для обсуждения того, что вы хотите изменить.

### Настройка разработки

1. Клонируйте репозиторий:

```bash
git clone https://github.com/ilyabrin/disk.git
cd disk
```

1. Установите зависимости:

```bash
go mod download
```

1. Запустите тесты:

```bash
go test -v
```

### Рекомендации

- Пишите тесты для новых функций
- Следуйте лучшим практикам и идиомам Go
- Обновляйте документацию при изменении API
- Убедитесь, что все тесты проходят перед отправкой PR

## 📄 Лицензия

Этот проект лицензирован под лицензией MIT - см. файл [LICENSE](./LICENSE) для подробностей.

## 🔗 Ссылки

- [Документация REST API Яндекс.Диска](https://yandex.ru/dev/disk/rest/)
- [Яндекс OAuth](https://oauth.yandex.ru/)
- [Репозиторий GitHub](https://github.com/ilyabrin/disk)

## 📝 Журнал изменений

### Последние обновления

- ✅ Всесторонняя поддержка пагинации с множеством стратегий
- ✅ Пакетные операции для эффективной обработки нескольких файлов
- ✅ Расширенный функционал загрузки с отслеживанием прогресса
- ✅ Улучшенная обработка ошибок и логирование
- ✅ Полная поддержка context.Context
- ✅ Улучшения безопасности с валидацией путей

## 💬 Поддержка

Если у вас есть вопросы или нужна помощь:

- Откройте [issue](https://github.com/ilyabrin/disk/issues)
- Проверьте существующие [примеры](./examples/)
- Прочитайте [документацию по пагинации](./PAGINATION.md)

## ⭐ Благодарности

Создано с ❤️ для Go-сообщества. Если эта библиотека вам помогла, пожалуйста, подумайте о том, чтобы поставить звезду на GitHub!

---

**Примечание:** Эта библиотека не является официально связанной с Яндексом. Это поддерживаемая сообществом клиентская библиотека для API Яндекс.Диска.
