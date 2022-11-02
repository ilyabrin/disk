package disk

// Disk Данные о свободном и занятом пространстве на Диске
type Disk struct {
	UnlimitedAutouploadEnabled bool           `json:"unlimited_autoupload_enabled,omitempty"`
	MaxFileSize                int            `json:"max_file_size,omitempty"`
	TotalSpace                 int            `json:"total_space,omitempty"`
	TrashSize                  int            `json:"trash_size,omitempty"`
	IsPaid                     bool           `json:"is_paid,omitempty"`
	UsedSpace                  int            `json:"used_space,omitempty"`
	SystemFolders              *SystemFolders `json:"system_folders,omitempty"`
	User                       *User          `json:"user,omitempty"`
	Revision                   int            `json:"revision,omitempty"`
}

// SystemFolders ...
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

// User ...
type User struct {
	Country     string `json:"country,omitempty"`
	Login       string `json:"login,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	UID         string `json:"uid,omitempty"`
}

// Resource ...
type Resource struct {
	AntivirusStatus  string            `json:"antivirus_status,omitempty"`
	ResourceID       string            `json:"resource_id,omitempty"`
	Share            *ShareInfo        `json:"share,omitempty"`
	File             string            `json:"file,omitempty"`
	Size             int               `json:"size,omitempty"`
	PhotosliceTime   string            `json:"photoslice_time,omitempty"`
	Embedded         *ResourceList     `json:"_embedded,omitempty"`
	Exif             *Exif             `json:"exif,omitempty"`
	CustomProperties map[string]string `json:"custom_properties,omitempty"`
	MediaType        string            `json:"media_type,omitempty"`
	Preview          string            `json:"preview,omitempty"`
	Type             string            `json:"type"`
	MimeType         string            `json:"mime_type,omitempty"`
	Revision         int               `json:"revision,omitempty"`
	PublicURL        string            `json:"public_url,omitempty"`
	Path             string            `json:"path"`
	Md5              string            `json:"md5,omitempty"`
	PublicKey        string            `json:"public_key,omitempty"`
	Sha256           string            `json:"sha256,omitempty"`
	Name             string            `json:"name"`
	Created          string            `json:"created"`
	Modified         string            `json:"modified"`
	CommentIDs       *CommentIds       `json:"comment_ids,omitempty"`
}

// PublicResource ...
type PublicResource struct {
	Resource
	Owner      *UserPublicInformation `json:"owner,omitempty"`
	ViewsCount int                    `json:"views_count,omitempty"`
}

// ShareInfo ...
type ShareInfo struct {
	IsRoot  bool   `json:"is_root,omitempty"`
	IsOwned bool   `json:"is_owned,omitempty"`
	Rights  string `json:"rights"`
}

// ResourceList ... Список ресурсов, содержащихся в папке. Содержит объекты Resource и свойства списка.
type ResourceList struct {
	Sort   string      `json:"sort,omitempty"`
	Items  []*Resource `json:"items"`
	Limit  int         `json:"limit,omitempty"`
	Offset int         `json:"offset,omitempty"`
	Path   string      `json:"path"`
	Total  int         `json:"total,omitempty"`
}

// Exif ...
type Exif struct {
	DateTime string `json:"date_time,omitempty"`
}

// CommentIds ...
type CommentIds struct {
	PrivateResource string `json:"private_resource,omitempty"`
	PublicResource  string `json:"public_resource,omitempty"`
}

// Link ...
type Link struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated,omitempty"`
}

// ResourceUploadLink ...
type ResourceUploadLink struct {
	OperationID string `json:"operation_id"`
	Href        string `json:"href"`
	Method      string `json:"method"`
	Templated   bool   `json:"templated,omitempty"`
}

// PublicResourcesList ...
type PublicResourcesList struct {
	Items  []*Resource `json:"items"`
	Type   string      `json:"type"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// LastUploadedResourceList ...
type LastUploadedResourceList struct {
	Items []*Resource `json:"items"`
	Limit int         `json:"limit,omitempty"`
}

// FilesResourceList ...
type FilesResourceList struct {
	Items  []*Resource `json:"items"`
	Limit  int         `json:"limit,omitempty"`
	Offset int         `json:"offset,omitempty"`
}

// UserPublicInformation ...
type UserPublicInformation struct {
	Login       string `json:"login,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	UID         string `json:"uid,omitempty"`
}

// Operation ...
type Operation struct {
	Status string `json:"status"`
}

// ErrorResponse ...
type ErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	StatusCode  int    `json:"status_code"`
	Error       error  `json:"error"` // TODO: []errors
}

// TrashResource ...
type TrashResource Resource
