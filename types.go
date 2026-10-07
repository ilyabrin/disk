package disk

// Disk describes the user's Disk as a whole: space, limits and the owner.
// [Client.DiskInfo] returns it.
//
// Sizes are in bytes.
type Disk struct {
	// UnlimitedAutouploadEnabled reports whether photos and videos uploaded
	// automatically from phones do not count against the space limit.
	UnlimitedAutouploadEnabled bool `json:"unlimited_autoupload_enabled,omitempty"`
	// MaxFileSize is the largest file that can be uploaded.
	MaxFileSize int64 `json:"max_file_size,omitempty"`
	// TotalSpace is the size of the Disk.
	TotalSpace int64 `json:"total_space,omitempty"`
	// TrashSize is how much the trash takes up. It counts towards UsedSpace.
	TrashSize int64 `json:"trash_size,omitempty"`
	// IsPaid reports whether the user has a paid plan.
	IsPaid bool `json:"is_paid,omitempty"`
	// UsedSpace is how much of TotalSpace is taken, trash included.
	UsedSpace int64 `json:"used_space,omitempty"`
	// SystemFolders holds the paths of folders the Disk manages itself.
	SystemFolders *SystemFolders `json:"system_folders,omitempty"`
	// User is the owner of the Disk.
	User *User `json:"user,omitempty"`
	// Revision changes whenever anything on the Disk changes. Yandex sends
	// it as a timestamp in microseconds.
	Revision int64 `json:"revision,omitempty"`
}

// SystemFolders holds the paths of the folders the Disk creates and fills on
// its own, such as downloads or photos from social networks. Each value is a
// path like "disk:/Загрузки/", in the language of the account.
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

// User is the owner of a Disk, as part of [Disk].
type User struct {
	// Country is the two-letter country code, such as "ru".
	Country string `json:"country,omitempty"`
	// Login is the Yandex login.
	Login string `json:"login,omitempty"`
	// DisplayName is the name shown to other people.
	DisplayName string `json:"display_name,omitempty"`
	// Uid is the numeric user ID, as a string.
	Uid string `json:"uid,omitempty"`
}

// Resource is a file or a folder on the Disk.
//
// For a folder fetched with its contents, Embedded holds one page of the
// items inside it. Times are ISO 8601 strings, such as
// "2026-10-07T12:00:00+00:00"; sizes are in bytes.
type Resource struct {
	// AntivirusStatus is the result of the virus check of a file, such as
	// "clean".
	AntivirusStatus string `json:"antivirus_status,omitempty"`
	// ResourceID identifies the resource and stays the same when it is moved
	// or renamed.
	ResourceID string `json:"resource_id,omitempty"`
	// Share is set when the resource is inside a shared folder.
	Share *ShareInfo `json:"share,omitempty"`
	// File is a link to download the file. Folders have none.
	File string `json:"file,omitempty"`
	// Size is the file size in bytes, zero for folders.
	Size int64 `json:"size,omitempty"`
	// PhotosliceTime is when a photo or video was taken, if the Disk knows it.
	PhotosliceTime string `json:"photoslice_time,omitempty"`
	// Embedded is one page of a folder's contents, when it was requested.
	Embedded *ResourceList `json:"_embedded,omitempty"`
	// Exif holds metadata read from a photo.
	Exif *Exif `json:"exif,omitempty"`
	// CustomProperties are the attributes set with [Client.UpdateMetadata].
	CustomProperties interface{} `json:"custom_properties,omitempty"`
	// MediaType is the kind of file as the Disk sees it, such as "image",
	// "video", "audio" or "document".
	MediaType string `json:"media_type,omitempty"`
	// Preview is a link to a thumbnail. Fetching it needs the access token.
	Preview string `json:"preview,omitempty"`
	// Type is "file" or "dir".
	Type string `json:"type"`
	// MimeType is the MIME type of a file, such as "image/jpeg".
	MimeType string `json:"mime_type,omitempty"`
	// Revision is the Disk revision in which the resource last changed.
	Revision int64 `json:"revision,omitempty"`
	// PublicURL is the public link, set only while the resource is published.
	PublicURL string `json:"public_url,omitempty"`
	// Path is the full path, such as "disk:/Photos/cat.jpg".
	Path string `json:"path"`
	// Md5 is the MD5 hash of a file's contents.
	Md5 string `json:"md5,omitempty"`
	// PublicKey identifies a published resource in the public API, such as
	// [Client.GetMetadataForPublicResource].
	PublicKey string `json:"public_key,omitempty"`
	// Sha256 is the SHA-256 hash of a file's contents.
	Sha256 string `json:"sha256,omitempty"`
	// Name is the last element of Path.
	Name string `json:"name"`
	// Created is when the resource was created.
	Created string `json:"created"`
	// Modified is when the resource last changed.
	Modified string `json:"modified"`
	// CommentIDs identify the comment threads of the resource.
	CommentIDs *CommentIds `json:"comment_ids,omitempty"`
}

// PublicResource is a published file or folder as seen by anyone with its
// link. The public API methods, such as
// [Client.GetMetadataForPublicResource], return it.
type PublicResource struct {
	Resource
	// Owner is the user who published the resource.
	Owner *UserPublicInformation `json:"owner,omitempty"`
	// ViewsCount is how many times the public link has been opened.
	ViewsCount int `json:"views_count,omitempty"`
}

// ShareInfo describes the user's access to a shared folder.
type ShareInfo struct {
	// IsRoot reports whether this resource is the shared folder itself
	// rather than something inside it.
	IsRoot bool `json:"is_root,omitempty"`
	// IsOwned reports whether the user owns the shared folder.
	IsOwned bool `json:"is_owned,omitempty"`
	// Rights is the user's access level, such as "r" or "rw".
	Rights string `json:"rights"`
}

// ResourceList is one page of a folder's contents, in [Resource.Embedded].
type ResourceList struct {
	// Sort is the field the items are sorted by, such as "name" or
	// "-modified" (a minus means descending).
	Sort string `json:"sort,omitempty"`
	// Items are the files and folders on this page.
	Items []*Resource `json:"items"`
	// Limit is the page size that was requested.
	Limit int `json:"limit,omitempty"`
	// Offset is how many items come before this page.
	Offset int `json:"offset,omitempty"`
	// Path is the folder being listed.
	Path string `json:"path"`
	// Total is the number of items in the folder, across all pages.
	Total int `json:"total,omitempty"`
}

// Exif holds metadata read from a photo.
type Exif struct {
	// DateTime is when the photo was taken, as an ISO 8601 string.
	DateTime string `json:"date_time,omitempty"`
}

// CommentIds identify the comment threads attached to a resource.
type CommentIds struct {
	// PrivateResource is the thread seen by people with access to the Disk.
	PrivateResource string `json:"private_resource,omitempty"`
	// PublicResource is the thread seen through the public link.
	PublicResource string `json:"public_resource,omitempty"`
}

// Link is a URL returned by the API: where to download a file, or, for an
// operation that runs in the background, where to check its status (see
// [Client.GetOperationStatus]).
type Link struct {
	// Href is the URL.
	Href string `json:"href"`
	// Method is the HTTP method to use with Href, such as "GET".
	Method string `json:"method"`
	// Templated reports whether Href contains placeholders to fill in.
	Templated bool `json:"templated,omitempty"`
}

// ResourceUploadLink is where to send a file's contents, returned by
// [Client.GetLinkForUpload]. The upload helpers in this package use it for
// you.
type ResourceUploadLink struct {
	// OperationID identifies the upload; its status can be checked with
	// [Client.GetOperationStatus].
	OperationID string `json:"operation_id"`
	// Href is the URL to send the contents to.
	Href string `json:"href"`
	// Method is the HTTP method to use, usually "PUT".
	Method string `json:"method"`
	// Templated reports whether Href contains placeholders to fill in.
	Templated bool `json:"templated,omitempty"`
}

// PublicResourcesList is one page of the user's published files and
// folders, returned by [Client.GetPublicResources].
type PublicResourcesList struct {
	// Items are the published resources on this page.
	Items []*Resource `json:"items"`
	// Type is the filter that was applied: "file", "dir" or empty for both.
	Type string `json:"type"`
	// Limit is the page size that was requested.
	Limit int `json:"limit"`
	// Offset is how many items come before this page.
	Offset int `json:"offset"`
}

// LastUploadedResourceList is the user's most recently uploaded files,
// returned by [Client.GetLastUploadedResources].
type LastUploadedResourceList struct {
	// Items are the files, newest first.
	Items []*Resource `json:"items"`
	// Limit is how many files were requested.
	Limit int `json:"limit,omitempty"`
}

// FilesResourceList is one page of all files on the Disk, regardless of
// folder, returned by [Client.GetSortedFiles].
type FilesResourceList struct {
	// Items are the files on this page.
	Items []*Resource `json:"items"`
	// Limit is the page size that was requested.
	Limit int `json:"limit,omitempty"`
	// Offset is how many files come before this page.
	Offset int `json:"offset,omitempty"`
}

// UserPublicInformation is what anyone can see about the owner of a
// published resource.
type UserPublicInformation struct {
	// Login is the Yandex login.
	Login string `json:"login,omitempty"`
	// DisplayName is the name shown to other people.
	DisplayName string `json:"display_name,omitempty"`
	// Uid is the numeric user ID, as a string.
	Uid string `json:"uid,omitempty"`
}

// Operation is the state of a background operation, such as copying a large
// folder, as reported by [Client.GetOperationStatus].
type Operation struct {
	// Status is "success", "failed" or "in-progress".
	Status string `json:"status"`
}

// ErrorResponse is an error reported by the API, or by this package when a
// request could not be made or its answer could not be read.
//
// Error is a short machine-readable code, such as "DiskNotFoundError";
// Description is in English and Message is in the language of the request.
type ErrorResponse struct {
	// Message is a description meant for people.
	Message string `json:"message"`
	// Description is a description in English.
	Description string `json:"description"`
	// Error is the error code, such as "DiskNotFoundError" or
	// "UnauthorizedError".
	Error string `json:"error"`

	// StatusCode is the HTTP status that produced this error, or 0 when the
	// request never reached a response (connection failure, decode error).
	// It is not part of the API payload; this package fills it in.
	StatusCode int `json:"-"`
}

// TrashResource is a file or folder in the trash.
type TrashResource struct {
	Resource
	// OriginPath is where the resource was before it was deleted, and where
	// restoring puts it back.
	OriginPath string `json:"origin_path,omitempty"`
	// Deleted is when the resource was moved to the trash.
	Deleted string `json:"deleted,omitempty"`
}

// TrashResourceList is one page of the trash, returned by
// [Client.ListTrashResources].
type TrashResourceList struct {
	// Items are the deleted resources on this page.
	Items []*TrashResource `json:"items"`
	// Limit is the page size that was requested.
	Limit int `json:"limit,omitempty"`
	// Offset is how many items come before this page.
	Offset int `json:"offset,omitempty"`
	// Path is the trash folder being listed, "trash:/" for the root.
	Path string `json:"path"`
	// Total is the number of items in the trash, across all pages.
	Total int `json:"total,omitempty"`
}
