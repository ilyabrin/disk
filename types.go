package disk

// GET /v1/disk
type Disk struct {
	UnlimitedAutouploadEnabled bool           `json:"unlimited_autoupload_enabled,omitempty"` // boolean, optional:
	MaxFileSize                int            `json:"max_file_size,omitempty"`                // integer, optional:
	TotalSpace                 int            `json:"total_space,omitempty"`                  // integer, optional:
	TrashSize                  int            `json:"trash_size,omitempty"`                   // integer, optional:
	IsPaid                     bool           `json:"is_paid,omitempty"`                      // boolean, optional:
	UsedSpace                  int            `json:"used_space,omitempty"`                   // integer, optional:
	SystemFolders              *SystemFolders `json:"system_folders,omitempty"`               // (SystemFolders, optional)
	User                       *User          `json:"user,omitempty"`                         // (User, optional)
	Revision                   int            `json:"revision,omitempty"`                     // (integer, optional):
}

type SystemFolders struct {
	Odnoklassniki string `json:"odnoklassniki,omitempty"`
	Google        string `json:"google,omitempty"`
	Instagram     string `json:"instagram,omitempty"`
	Vkontakte     string `json:"vkontakte,omitempty"`
	Mailru        string `json:"mailru,omitempty"`
	Downloads     string `json:"downloads,omitempty"`
	Applications  string `json:"applications,omitempty"`
	Facebook      string `json:"facebook,omitempty"`
	Social        string `json:"social,omitempty"`
	Screenshots   string `json:"screenshots,omitempty"`
	Photostream   string `json:"photostream,omitempty"`
}

type User struct {
	Country     string `json:"country,omitempty"`      // string, optional: <Страна>,
	Login       string `json:"login,omitempty"`        // string, optional: <Логин>,
	DisplayName string `json:"display_name,omitempty"` // string, optional: <Отображаемое имя>,
	Uid         string `json:"uid,omitempty"`          // string, optional: <Идентификатор пользователя>
}

type Resource struct {
	AntivirusStatus  string            `json:"antivirus_status,omitempty"`  // (object, optional): <Статус проверки антивирусом>,
	ResourceID       string            `json:"resource_id,omitempty"`       // (string, optional): <Идентификатор ресурса>,
	Share            *ShareInfo        `json:"share,omitempty"`             // (ShareInfo, optional),
	File             string            `json:"file,omitempty"`              // (string, optional): <URL для скачивания файла>,
	Size             int               `json:"size,omitempty"`              // (integer, optional): <Размер файла>,
	PhotosliceTime   string            `json:"photoslice_time,omitempty"`   // (string, optional): <Дата создания фото или видео файла>,
	Embedded         *ResourceList     `json:"_embedded,omitempty"`         // (ResourceList, optional),
	Exif             *Exif             `json:"exif,omitempty"`              // (Exif, optional),
	CustomProperties map[string]string `json:"custom_properties,omitempty"` // (object, optional): <Пользовательские атрибуты ресурса>,
	MediaType        string            `json:"media_type,omitempty"`        // (string, optional): <Определённый Диском тип файла>,
	Preview          string            `json:"preview,omitempty"`           // (string, optional): <URL превью файла>,
	Type             string            `json:"type"`                        // (string): <Тип>,
	MimeType         string            `json:"mime_type,omitempty"`         // (string, optional): <MIME-тип файла>,
	Revision         int               `json:"revision,omitempty"`          // (integer, optional): <Ревизия Диска в которой этот ресурс был изменён последний раз>,
	PublicURL        string            `json:"public_url,omitempty"`        // (string, optional): <Публичный URL>,
	Path             string            `json:"path"`                        // (string): <Путь к ресурсу>,
	Md5              string            `json:"md5,omitempty"`               // (string, optional): <MD5-хэш>,
	PublicKey        string            `json:"public_key,omitempty"`        // (string, optional): <Ключ опубликованного ресурса>,
	Sha256           string            `json:"sha256,omitempty"`            // (string, optional): <SHA256-хэш>,
	Name             string            `json:"name"`                        // (string): <Имя>,
	Created          string            `json:"created"`                     // (string): <Дата создания>,
	Modified         string            `json:"modified"`                    // (string): <Дата изменения>,
	CommentIDs       *CommentIds       `json:"comment_ids,omitempty"`       // (CommentIds, optional)
}

type PublicResource struct {
	Resource
	Owner      *UserPublicInformation `json:"owner,omitempty"`
	ViewsCount int                    `json:"views_count,omitempty"`
}

type ShareInfo struct {
	IsRoot  bool   `json:"is_root,omitempty"`
	IsOwned bool   `json:"is_owned,omitempty"`
	Rights  string `json:"rights"`
}

type ResourceList struct {
	Sort   string      `json:"sort,omitempty"`   //  (string, optional): <Поле, по которому отсортирован список>,
	Items  []*Resource `json:"items"`            //  (Array[Resource]): <Элементы списка>,
	Limit  int         `json:"limit,omitempty"`  //  (integer, optional): <Количество элементов на странице>,
	Offset int         `json:"offset,omitempty"` //  (integer, optional): <Смещение от начала списка>,
	Path   string      `json:"path"`             //  (string): <Путь к ресурсу, для которого построен список>,
	Total  int         `json:"total,omitempty"`  //  (integer, optional): <Общее количество элементов в списке>}
}

type Exif struct {
	DateTime string `json:"date_time,omitempty"`
}

type CommentIds struct {
	PrivateResource string `json:"private_resource,omitempty"`
	PublicResource  string `json:"public_resource,omitempty"`
}

type Link struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated,omitempty"`
}

type ResourceUploadLink struct {
	OperationID string `json:"operation_id"`        // (string): <Идентификатор операции загрузки файла>,
	Href        string `json:"href"`                // (string): <URL>,
	Method      string `json:"method"`              // (string): <HTTP-метод>,
	Templated   bool   `json:"templated,omitempty"` // (boolean, optional): <Признак шаблонизированного URL>
}

type PublicResourcesList struct {
	Items  []*Resource `json:"items"`  // (Array[Resource]): <Элементы списка>,
	Type   string      `json:"type"`   // (string, optional): <Значение фильтра по типу ресурсов>,
	Limit  int         `json:"limit"`  // (integer, optional): <Количество элементов на странице>,
	Offset int         `json:"offset"` // (integer, optional): <Смещение от начала списка>
}

type LastUploadedResourceList struct {
	Items []*Resource `json:"items"`           //(Array[Resource]): <Элементы списка>,
	Limit int         `json:"limit,omitempty"` // (integer, optional): <Количество элементов на странице>
}

type FilesResourceList struct {
	Items  []*Resource `json:"items"`            // (Array[Resource]): <Элементы списка>,
	Limit  int         `json:"limit,omitempty"`  //(integer, optional): <Количество элементов на странице>,
	Offset int         `json:"offset,omitempty"` // (integer, optional): <Смещение от начала списка>
}

type UserPublicInformation struct {
	Login       string `json:"login,omitempty"`        // (string, optional): <Логин.>,
	DisplayName string `json:"display_name,omitempty"` // (string, optional): <Отображаемое имя пользователя.>,
	Uid         string `json:"uid,omitempty"`          // (string, optional): <Идентификатор пользователя.>
}

type Operation struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	StatusCode  int    `json:"status_code"`
	Error       string `json:"error"`
}

type TrashResource Resource
