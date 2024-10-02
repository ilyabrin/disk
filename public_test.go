package disk

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetadataForPublicResource(t *testing.T) {
	// create http handler
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

		w.Write([]byte(`
		{
			"antivirus_status": "clean",
			"views_count": 120,
			"resource_id": "1:longhash",
			"file": "https://downloader.disk.yandex.ru/disk/hash/123/xyz&filename=file.zip",
			"owner": {
			  "login": "username",
			  "display_name": "Ilya",
			  "uid": "1"
			},
			"size": 123456789,
			"photoslice_time": "2020-01-14T12:21:46+00:00",
			"exif": {
			  "date_time": "2020-01-14T12:45:46+00:00"
			},
			"media_type": "video",
			"preview": "https://downloader.disk.yandex.ru/disk/hash/123/xyz&filename=file.zip",
			"type": "file",
			"mime_type": "video/quicktime",
			"revision": 1234567898765432,
			"public_url": "https://yadi.sk/i/xXxqcxV1mOA123",
			"path": "/",
			"md5": "123",
			"public_key": "123+cfrt/bbb+q/453==",
			"sha256": "123",
			"name": "file.zip",
			"created": "2020-01-10T07:07:07+00:00",
			"sizes": [
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XXXS"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XXS"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XS"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "S"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "M"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "L"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XL"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XXL"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "XXXL"
			  },
			  {
				"url": "https://downloader.disk.yandex.ru/preview/123/xxx/abc?uid=0&filename=file.zip",
				"name": "C"
			  }
			],
			"modified": "2020-01-13T07:07:07+00:00",
			"comment_ids": {
			  "private_resource": "1:123",
			  "public_resource": "1:123"
			}
		  }`))
	})

	client := mockedHttpClient(h)

	metadata, err := client.GetMetadataForPublicResource(context.Background(), "https://disk.yandex.ru/i/123ABC-321cba")

	// check response data
	assert.Nil(t, err)
	assert.Equal(t, "file.zip", metadata.Name)

	// check type
	assert.IsType(t, &PublicResource{}, metadata)
}

func TestGetDownloadURLForPublicResource(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

		w.Write([]byte(`
        {
			"href": "https://downloader.disk.yandex.ru/disk/123/abc/hash&filename=file.zip",
			"method": "GET",
			"templated": false
		}`))
	})

	client := mockedHttpClient(h)

	link, err := client.GetDownloadURLForPublicResource(context.Background(), "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	// check response data
	assert.Nil(t, err)
	assert.Equal(t, "https://downloader.disk.yandex.ru/disk/123/abc/hash&filename=file.zip", link.Href)

	// check type
	assert.IsType(t, &Link{}, link)
}

func TestSavePublicResource(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

		w.Write([]byte(`
        {
            "href": "https://cloud-api.yandex.net/v1/disk/resources?path=file.zip",
            "method": "GET",
            "templated": false
        }`))
	})

	client := mockedHttpClient(h)
	link, err := client.SavePublicResource(context.Background(), "https://disk.yandex.ru/i/12_xfKBSSOnf21")

	// check response data
	assert.Nil(t, err)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk/resources?path=file.zip", link.Href)

	// check type
	assert.IsType(t, &Link{}, link)
}
