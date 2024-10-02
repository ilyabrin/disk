package disk

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientGetDiskInfo(t *testing.T) {

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

		w.Write([]byte(`
		{
			"unlimited_autoupload_enabled": false,
			"max_file_size": 53687091200,
			"total_space": 1190242811904,
			"is_paid": true,
			"used_space": 664410431972,
			"system_folders": {
				"odnoklassniki": "disk:/Социальные сети/Одноклассники",
				"google": "disk:/Социальные сети/Google+",
				"instagram": "disk:/Социальные сети/Instagram",
				"vkontakte": "disk:/Социальные сети/ВКонтакте",
				"mailru": "disk:/Социальные сети/Мой Мир",
				"downloads": "disk:/Загрузки/",
				"applications": "disk:/Приложения",
				"facebook": "disk:/Социальные сети/Facebook",
				"social": "disk:/Социальные сети/",
				"screenshots": "disk:/Скриншоты/",
				"photostream": "disk:/Фотокамера/"
			   },
			"user": {
			 "country": "ru",
			 "login": "user",
			 "display_name": "User Name",
			 "uid": "12345678"
			},
			"revision": 1602851010832695
		}`))
	})

	client := mockedHttpClient(h)

	disk, err := client.DiskInfo(context.Background())

	// check response data
	assert.Nil(t, err)
	assert.Equal(t, true, disk.IsPaid)
	assert.Equal(t, 53687091200, disk.MaxFileSize)
	assert.Equal(t, 1190242811904, disk.TotalSpace)
	assert.Equal(t, "User Name", disk.User.DisplayName)

	// check types
	assert.IsType(t, Disk{}.IsPaid, disk.IsPaid)
	assert.IsType(t, Disk{}.User, disk.User)
	assert.IsType(t, Disk{}.SystemFolders, disk.SystemFolders)
}
