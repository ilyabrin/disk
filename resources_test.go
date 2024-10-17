package disk

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDeleteResourceURL(t *testing.T) {
	// TODO
	client := &Client{}

	tests := []struct {
		name        string
		path        string
		permanently bool
		want        string
	}{
		{
			name:        "Delete temporary file",
			path:        "/path/to/file.txt",
			permanently: false,
			want:        "resources?path=%2Fpath%2Fto%2Ffile.txt&permanent=false",
		},
		{
			name:        "Delete permanent file",
			path:        "/another/path/to/file.jpg",
			permanently: true,
			want:        "resources?path=%2Fanother%2Fpath%2Fto%2Ffile.jpg&permanent=true",
		},
		{
			name:        "Delete file with special characters",
			path:        "/path with spaces/file with &.txt",
			permanently: false,
			want:        "resources?path=%2Fpath+with+spaces%2Ffile+with+%26.txt&permanent=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildDeleteResourceURL(tt.path, tt.permanently)
			if got != tt.want {
				t.Errorf("buildDeleteResourceURL() = %v, want %v", got, tt.want)
			}

			// Additional check: ensure the generated URL is valid
			_, err := url.Parse(got)
			if err != nil {
				t.Errorf("buildDeleteResourceURL() generated an invalid URL: %v", err)
			}
		})
	}
}

// todo: add *ErrorResponse to return
// todo: add *http.Response to return
func TestDeleteResource(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`no content`))
		}))

	err := client.DeleteResource(context.Background(), "testdir2", true)

	assert.Nil(t, err)
}

func TestGetMetadata(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"_embedded": {
  					  "sort": "",
  					  "items": [],
  					  "limit": 20,
  					  "offset": 0,
  					  "path": "disk:/testdir",
  					  "total": 0
  					},
  					"name": "testdir",
  					"exif": {},
  					"resource_id": "123123:15a9a26c342e6f64X8e4b9d02cC8es0c4db1eb70678d7e956dfb2923ee886bc5",
  					"created": "2024-10-14T17:10:00+00:00",
  					"modified": "2024-10-14T17:10:00+00:00",
  					"path": "disk:/testdir",
  					"comment_ids": {
  					  "private_resource": "1213123:15a9a26c342e6f64X8e4b9d02cC8es0c4db1eb70678d7e956dfb2923ee886bc5",
  					  "public_resource": "123123:15a9a26c342e6f64X8e4b9d02cC8es0c4db1eb70678d7e956dfb2923ee886bc5"
  					},
  					"type": "dir",
  					"revision": 1234567894721812
				}`))
		}))

	disk, _ := client.GetMetadata(context.Background(), "testdir")

	assert.IsType(t, &Resource{}, disk)
}

/*
todo: add examples to README

	newMeta := map[string]map[string]string{
		"custom_properties": {
			"key_01": "value_01",
			"key_02": "value_02",
			"key_07": "value_07",
		},
	}
*/
func TestUpdateMetadata(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
					"antivirus_status": "clean",
					"resource_id": "string",
					"share": {
						"is_root": true,
						"is_owned": true,
						"rights": "string"
					},
					"file": "string",
					"size": 0,
					"photoslice_time": "2024-10-17T18:15:12.282Z",
					"_embedded": {
						"sort": "string",
						"items": [{}],
						"limit": 0,
						"offset": 0,
						"path": "string",
						"total": 0
					},
					"exif": {
						"date_time": "2024-10-17T18:15:12.282Z",
						"gps_longitude": {},
						"gps_latitude": {}
					},
					"custom_properties": {
						"key_01": "value_01",
						"key_02": "value_02",
						"key_07": "value_07"
					},
					"media_type": "string",
					"preview": "string",
					"type": "string",
					"mime_type": "string",
					"revision": 0,
					"public_url": "string",
					"path": "string",
					"md5": "string",
					"public_key": "string",
					"sha256": "string",
					"name": "string",
					"created": "2024-10-17T18:15:12.282Z",
					"sizes": [{
						"url": "string",
						"name": "string"
					}],
					"modified": "2024-10-17T18:15:12.283Z",
					"comment_ids": {
						"private_resource": "string",
						"public_resource": "string"
					}
				}`))
		}))

	// TODO: move to CustomProperty type
	newMeta := map[string]map[string]string{"custom_properties": {
		"key_01": "value_01",
		"key_02": "value_02",
		"key_07": "value_07",
	}}
	resource, err := client.UpdateMetadata(context.Background(), "testdir2", newMeta)

	assert.Nil(t, err)
	assert.IsType(t, &Resource{}, resource)
	// TODO: change type from 'string' to 'map[string]map[string]string{}'
	// assert.IsType(t, []CustomProperty, resource.CustomProperties)

}

// CreateDir creates a new directory with the specified 'path' name.
// todo: can't create nested dirs like newDir/subDir/anotherDir
func TestCreateDir(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(`
        	{
  				"href": "string",
  				"method": "string",
  				"templated": true
			}`))
		}))

	link, err := client.CreateDir(context.Background(), "testdir")

	assert.IsType(t, &Link{}, link)
	assert.IsType(t, &ErrorResponse{}, err)

}

func TestCopyResource(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
					"href": "https://cloud-api.yandex.net/v1/disk/resources?path=disk%3A%2Ftestdir2",
					"method": "GET",
					"templated": false
				}`))
		}))

	disk, _ := client.CopyResource(context.Background(), "testdir", "testdir2")

	assert.Equal(t, "GET", disk.Method)
}

func TestGetDownloadURL(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"href": "https://downloader.disk.yandex.ru/zip/99a88a/1234670ee/ABCDc2df111dddp123==?uid=1&filename=testdir.zip&disposition=attachment&hash=&limit=0&owner_uid=1&tknv=v2",
  					"method": "GET",
  					"templated": false
				}`))
		}))

	disk, err := client.GetDownloadURL(context.Background(), "testdir")

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &Link{}, disk)
	assert.Equal(t, "GET", disk.Method)
}

func TestGetSortedFiles(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
					"items": [
						{
						"antivirus_status": "clean",
						"size": 2882112,
						"comment_ids": {
							"private_resource": "1234567890:088BhddcXfvnbb3Hd74",
							"public_resource": "1234567890:088BhddcXfvnbb3Hd74"
						},
						"name": "Book.pdf",
						"exif": {},
						"created": "2012-01-10T14:10:23+00:00",
						"resource_id": "1234567890:083BhddcXfvnbb3Hd74",
						"modified": "2012-01-10T14:10:23+00:00",
						"mime_type": "application/pdf",
						"sizes": [
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2",
							"name": "DEFAULT"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XXXS&crop=0",
							"name": "XXXS"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XXS&crop=0",
							"name": "XXS"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XS&crop=0",
							"name": "XS"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=S&crop=0",
							"name": "S"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=M&crop=0",
							"name": "M"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=L&crop=0",
							"name": "L"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XL&crop=0",
							"name": "XL"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XXL&crop=0",
							"name": "XXL"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=XXXL&crop=0",
							"name": "XXXL"
							},
							{
							"url": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=S&crop=0",
							"name": "C"
							}
						],
						"file": "https://downloader.disk.yandex.ru/disk/abcdf/abcd3/abcdf34123%3D%3D?uid=1234567890&filename=Book.pdf&disposition=attachment&hash=&limit=0&content_type=application%2Fpdf&owner_uid=1234567890&fsize=2882112&hid=bdd02a4b304ef7709e0c16e0890c867b&media_type=document&tknv=v2&etag=219d5b6e0b5a90ffa95b79db1d3c8aa7",
						"media_type": "document",
						"preview": "https://downloader.disk.yandex.ru/preview/abcdfx01/inf/abc-zxabdfe-abcd%3D%3D?uid=1234567890&filename=Book.pdf&disposition=inline&hash=&limit=0&content_type=image%2Fjpeg&owner_uid=1234567890&tknv=v2&size=S&crop=0",
						"path": "disk:/Books/Book.pdf",
						"sha256": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
						"type": "file",
						"md5": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
						"revision": 121312412345
						}
					],
					"limit": 1,
					"offset": 0
				}`))
		}))

	disk, err := client.GetSortedFiles(context.Background())

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &FilesResourceList{}, disk)
}

// get | sortBy = [name = default, uploadDate]
func TestGetLastUploadedResources(t *testing.T) {
	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
				"items": [
					{
					"antivirus_status": "clean",
					"resource_id": "string",
					"share": {
						"is_root": true,
						"is_owned": true,
						"rights": "string"
					},
					"file": "string",
					"size": 0,
					"photoslice_time": "2024-10-07T10:10:00.000Z",
					"_embedded": {
						"sort": "string",
						"items": [
						{}
						],
						"limit": 0,
						"offset": 0,
						"path": "string",
						"total": 0
					},
					"exif": {
						"date_time": "2024-10-07T10:10:00.000Z",
						"gps_longitude": {},
						"gps_latitude": {}
					},
					"custom_properties": {},
					"media_type": "string",
					"preview": "string",
					"type": "string",
					"mime_type": "string",
					"revision": 0,
					"public_url": "string",
					"path": "string",
					"md5": "string",
					"public_key": "string",
					"sha256": "string",
					"name": "string",
					"created": "2024-10-07T10:10:00.000Z",
					"sizes": [
						{
						"url": "string",
						"name": "string"
						}
					],
					"modified": "2024-10-07T10:10:00.000Z",
					"comment_ids": {
						"private_resource": "string",
						"public_resource": "string"
					}
					}
				],
				"limit": 0
				}`))
		}))

	disk, err := client.GetLastUploadedResources(context.Background())

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &LastUploadedResourceList{}, disk)
}

func TestMoveResource(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"href": "string",
  					"method": "string",
  					"templated": true
				}`))
		}))

	disk, err := client.MoveResource(context.Background(), "testdir/testfile", "testdir2")

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &Link{}, disk)
}

func TestGetPublicResources(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"items": [
  					  {
  					    "antivirus_status": "clean",
  					    "resource_id": "string",
  					    "share": {
  					      "is_root": true,
  					      "is_owned": true,
  					      "rights": "string"
  					    },
  					    "file": "string",
  					    "size": 0,
  					    "photoslice_time": "2024-10-07T21:15:04.117Z",
  					    "_embedded": {
  					      "sort": "string",
  					      "items": [
  					        {}
  					      ],
  					      "limit": 0,
  					      "offset": 0,
  					      "path": "string",
  					      "total": 0
  					    },
  					    "exif": {
  					      "date_time": "2024-10-07T21:15:04.117Z",
  					      "gps_longitude": {},
  					      "gps_latitude": {}
  					    },
  					    "custom_properties": {},
  					    "media_type": "string",
  					    "preview": "string",
  					    "type": "string",
  					    "mime_type": "string",
  					    "revision": 0,
  					    "public_url": "string",
  					    "path": "string",
  					    "md5": "string",
  					    "public_key": "string",
  					    "sha256": "string",
  					    "name": "string",
  					    "created": "2024-10-07T21:15:04.117Z",
  					    "sizes": [
  					      {
  					        "url": "string",
  					        "name": "string"
  					      }
  					    ],
  					    "modified": "2024-10-07T21:15:04.117Z",
  					    "comment_ids": {
  					      "private_resource": "string",
  					      "public_resource": "string"
  					    }
  					  }
  					],
  					"type": "string",
  					"limit": 0,
  					"offset": 0
				}`))
		}))

	disk, err := client.GetPublicResources(context.Background())

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &PublicResourcesList{}, disk)
}

func TestPublishResource(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"href": "string",
  					"method": "string",
  					"templated": true
				}`))
		}))

	disk, err := client.PublishResource(context.Background(), "testdir")

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &Link{}, disk)
}

func TestUnpublishResource(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
  					"href": "string",
  					"method": "string",
  					"templated": true
				}`))
		}))

	disk, err := client.UnpublishResource(context.Background(), "testdir")

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &Link{}, disk)
}

func TestGetLinkForUpload(t *testing.T) {

	client := mockedHttpClient(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, r.Header.Get("Authorization"))
			assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

			w.Write([]byte(
				`{
					"operation_id": "string",
					"href": "string",
					"method": "string",
					"templated": true
	  			}`))
		}))

	disk, err := client.GetLinkForUpload(context.Background(), "testdir")

	assert.IsType(t, &ErrorResponse{}, err)
	assert.IsType(t, &ResourceUploadLink{}, disk)
}

// todo: empty responses - fix it
func TestUploadFile(t *testing.T) {}
