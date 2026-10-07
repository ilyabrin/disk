package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// validatePath sanitizes and validates Yandex Disk resource paths to prevent
// path traversal. Disk paths are always POSIX-style, so this deliberately uses
// POSIX semantics rather than filepath (which would treat "\" as a separator
// on Windows).
func validatePath(p string) error {
	if p == "" {
		return errors.New("path cannot be empty")
	}

	// Remove any null bytes
	if strings.Contains(p, "\x00") {
		return errors.New("path contains null bytes")
	}

	// Check for excessively long paths
	if len(p) > 4096 {
		return errors.New("path too long")
	}

	// Reject traversal only when ".." is a whole path segment, so that regular
	// names such as "report..final.pdf" stay valid. The raw path is inspected
	// rather than the cleaned one: path.Clean would silently resolve "/a/../b"
	// into "/b" instead of flagging it.
	for _, segment := range strings.Split(p, "/") {
		if segment == ".." {
			return errors.New("path traversal detected")
		}
	}

	return nil
}

func (c *Client) buildDeleteResourceURL(path string, permanently bool) string {
	return c.deleteURL(path, &DeleteOptions{Permanently: permanently})
}

func (c *Client) deleteURL(path string, opts *DeleteOptions) string {
	if opts == nil {
		opts = &DeleteOptions{}
	}
	query := url.Values{}
	query.Set("path", path)
	query.Set("permanently", strconv.FormatBool(opts.Permanently))
	if opts.ForceAsync {
		query.Set("force_async", "true")
	}
	if opts.MD5 != "" {
		query.Set("md5", opts.MD5)
	}
	return fmt.Sprintf("resources?%s", query.Encode())
}

// DeleteOptions are the optional parameters of
// [Client.DeleteResourceWithOptions].
type DeleteOptions struct {
	// Permanently removes the resource for good instead of moving it to the
	// trash.
	Permanently bool
	// ForceAsync makes the API delete in the background even when the
	// resource is small.
	ForceAsync bool
	// MD5, when set, deletes a file only if its contents still have this
	// hash, so a file changed in the meantime is kept. Files only.
	MD5 string
}

// DeleteResource deletes the file or folder at path. It goes to the trash,
// or is removed for good when permanently is true.
func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	return c.DeleteResourceWithOptions(ctx, path, &DeleteOptions{Permanently: permanently})
}

// DeleteResourceWithOptions is [Client.DeleteResource] with the choice to
// run in the background or to delete a file only if it is unchanged.
func (c *Client) DeleteResourceWithOptions(ctx context.Context, path string, opts *DeleteOptions) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("delete error: %w", err)
	}

	url := c.deleteURL(path, opts)

	resp, err := c.doRequest(ctx, DELETE, url, nil)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// The API answers 204 once the resource is gone, and 202 when a large
	// folder is being deleted in the background. 200 is kept for proxies.
	if _, err := c.handleResponse(resp, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}); err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}

	return nil
}

// ResourceOptions contains optional parameters for GetMetadata requests.
// When listing directory contents, Limit/Offset/Sort control pagination of the
// _embedded.items list returned by the API.
type ResourceOptions struct {
	Limit       int      // Number of items to return in _embedded (0 = API default)
	Offset      int      // Offset into the _embedded list
	Sort        string   // Sort field: "name", "path", "created", "modified", "size" (prefix with "-" for descending)
	PreviewSize string   // Thumbnail size, e.g. "S", "M", "L", "XL", "XXL", "XXXL" or "NxM"
	PreviewCrop bool     // Whether to crop preview to square
	Fields      []string // Response fields to return, e.g. "name", "_embedded.items.path" (empty = all)
}

func (o *ResourceOptions) apply(query url.Values) {
	if o == nil {
		return
	}
	if o.Limit > 0 {
		query.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		query.Set("offset", strconv.Itoa(o.Offset))
	}
	if o.Sort != "" {
		query.Set("sort", o.Sort)
	}
	if o.PreviewSize != "" {
		query.Set("preview_size", o.PreviewSize)
	}
	if o.PreviewCrop {
		query.Set("preview_crop", "true")
	}
	if len(o.Fields) > 0 {
		query.Set("fields", strings.Join(o.Fields, ","))
	}
}

// GetMetadata returns the file or folder at path. For a folder, Embedded
// holds the first page of its contents; use [Client.GetMetadataWithOptions]
// to choose the page, the order or the fields.
func (c *Client) GetMetadata(ctx context.Context, path string) (*Resource, *ErrorResponse) {
	return c.GetMetadataWithOptions(ctx, path, nil)
}

// GetMetadataWithOptions retrieves metadata for a file or directory.
// For directories the response includes an _embedded field with paginated contents.
// Use opts to control pagination (Limit/Offset), sorting and returned fields.
func (c *Client) GetMetadataWithOptions(ctx context.Context, path string, opts *ResourceOptions) (*Resource, *ErrorResponse) {
	if err := validatePath(path); err != nil {
		return nil, &ErrorResponse{Error: err.Error()}
	}

	query := url.Values{}
	query.Set("path", path)
	opts.apply(query)

	return requestJSON[Resource](ctx, c, GET, "resources?"+query.Encode(), nil)
}

// UpdateMetadata sets custom attributes on the file or folder at path and
// returns it with [Resource.CustomProperties] updated. The map is sent as
// the request body, so the attributes go under "custom_properties":
//
//	props := map[string]map[string]string{
//		"custom_properties": {"project": "yad", "reviewed": "yes"},
//	}
//	res, errResp := client.UpdateMetadata(ctx, "disk:/report.pdf", props)
func (c *Client) UpdateMetadata(ctx context.Context, path string, custom_properties map[string]map[string]string) (*Resource, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	body, err := json.Marshal(custom_properties)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to marshal properties: %v", err)}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Resource](ctx, c, PATCH, "resources?"+query.Encode(), bytes.NewBuffer(body))
}

// CreateDir creates a new directory with the specified 'path' name.
//
// Only the final segment is created: the parent must already exist, mirroring
// os.Mkdir. Use CreateDirAll to create a nested path in one call.
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources?"+query.Encode(), nil, http.StatusCreated)
}

// AlreadyExists reports whether this error means the resource is already
// present, which CreateDirAll treats as success rather than failure.
func (e *ErrorResponse) AlreadyExists() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusConflict &&
		e.Error == "DiskPathPointsToExistentDirectoryError"
}

// CreateDirAll creates path along with any missing parent directories,
// mirroring os.MkdirAll. Directories that already exist are left alone and do
// not produce an error; if path already exists as a directory, CreateDirAll
// does nothing and returns nil.
//
// The Yandex Disk API creates one level per request, so this issues one
// request per missing segment, walking from the shallowest to the deepest.
func (c *Client) CreateDirAll(ctx context.Context, path string) *ErrorResponse {
	if len(path) < 1 {
		return &ErrorResponse{Error: "path cannot be empty"}
	}
	if err := validatePath(path); err != nil {
		return &ErrorResponse{Error: err.Error()}
	}

	// "disk:/a/b" and "/a/b" and "a/b" all address the same place; normalise
	// to the segments so the prefixes below rebuild a well-formed path.
	trimmed := strings.TrimPrefix(path, "disk:")
	absolute := strings.HasPrefix(trimmed, "/")

	var segments []string
	for _, segment := range strings.Split(trimmed, "/") {
		if segment != "" && segment != "." {
			segments = append(segments, segment)
		}
	}
	if len(segments) == 0 {
		return &ErrorResponse{Error: "path contains no directory names"}
	}

	prefix := ""
	if absolute {
		prefix = "/"
	}

	for i, segment := range segments {
		if i > 0 {
			prefix += "/"
		}
		prefix += segment

		if _, errResp := c.CreateDir(ctx, prefix); errResp != nil && !errResp.AlreadyExists() {
			return errResp
		}
	}

	return nil
}

// CopyResource copies the file or folder at from to path. It fails if path
// already exists. Copying a large folder may continue in the background; the
// returned [Link] then points to the operation, see [Client.GetOperationStatus].
func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	return c.CopyResourceWithOptions(ctx, from, path, nil)
}

// CopyMoveOptions are the optional parameters of
// [Client.CopyResourceWithOptions] and [Client.MoveResourceWithOptions].
type CopyMoveOptions struct {
	// Overwrite replaces whatever is at the target path. Without it, an
	// existing target makes the request fail with 409.
	Overwrite bool
	// ForceAsync makes the API run the operation in the background even
	// when it is small, and return a link to it.
	ForceAsync bool
}

func (o *CopyMoveOptions) apply(query url.Values) {
	if o == nil {
		return
	}
	if o.Overwrite {
		query.Set("overwrite", "true")
	}
	if o.ForceAsync {
		query.Set("force_async", "true")
	}
}

// CopyResourceWithOptions is [Client.CopyResource] with the choice to
// overwrite the target or to run in the background.
func (c *Client) CopyResourceWithOptions(ctx context.Context, from, path string, opts *CopyMoveOptions) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)
	opts.apply(query)

	return requestJSON[Link](ctx, c, POST, "resources/copy?"+query.Encode(), nil,
		http.StatusOK, http.StatusCreated, http.StatusAccepted)
}

// GetDownloadURL returns a link to download the file at path. The link
// works without the access token for a limited time. For a folder, it
// downloads a zip archive.
func (c *Client) GetDownloadURL(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, GET, "resources/download?"+query.Encode(), nil)
}

// FilesOptions contains the filters accepted by the flat file list endpoint
// (/v1/disk/resources/files) in addition to pagination.
type FilesOptions struct {
	// MediaType filters by file category: "audio", "backup", "book",
	// "compressed", "data", "development", "diskimage", "document",
	// "encoded", "executable", "flash", "font", "image", "settings",
	// "spreadsheet", "text", "unknown", "video", "web".
	MediaType   []string
	Sort        string // "name", "path", "created", "modified", "size" (prefix "-" to reverse)
	PreviewSize string // Thumbnail size, e.g. "M" or "120x240"
	PreviewCrop bool   // Crop previews to the requested size
	Fields      []string
}

func (o *FilesOptions) apply(query url.Values) {
	if o == nil {
		return
	}
	if len(o.MediaType) > 0 {
		query.Set("media_type", strings.Join(o.MediaType, ","))
	}
	if o.Sort != "" {
		query.Set("sort", o.Sort)
	}
	if o.PreviewSize != "" {
		query.Set("preview_size", o.PreviewSize)
	}
	if o.PreviewCrop {
		query.Set("preview_crop", "true")
	}
	if len(o.Fields) > 0 {
		query.Set("fields", strings.Join(o.Fields, ","))
	}
}

// GetSortedFiles returns the first page of all files on the Disk, regardless
// of folder, in the API's default order. See
// [Client.GetSortedFilesWithOptions] to filter, sort or page through them.
func (c *Client) GetSortedFiles(ctx context.Context) (*FilesResourceList, *ErrorResponse) {
	return c.GetSortedFilesWithPagination(ctx, nil)
}

// GetSortedFilesWithPagination gets a sorted list of files with pagination support
func (c *Client) GetSortedFilesWithPagination(ctx context.Context, options *PaginationOptions) (*FilesResourceList, *ErrorResponse) {
	return c.GetSortedFilesWithOptions(ctx, options, nil)
}

// GetSortedFilesWithOptions returns the flat file list with pagination plus the
// media_type/sort/preview/fields filters supported by the API.
func (c *Client) GetSortedFilesWithOptions(ctx context.Context, options *PaginationOptions, filters *FilesOptions) (*FilesResourceList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	query := url.Values{}
	addPaginationParams(query, options)
	filters.apply(query)

	endpoint := "resources/files"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	return requestJSON[FilesResourceList](ctx, c, GET, endpoint, nil)
}

// GetSortedFilesPaged returns a paginated wrapper with pagination info
func (c *Client) GetSortedFilesPaged(ctx context.Context, options *PaginationOptions) (*PagedFilesResourceList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	files, errResp := c.GetSortedFilesWithPagination(ctx, options)
	if errResp != nil {
		return nil, errResp
	}

	// Create pagination info
	itemCount := len(files.Items)
	paginationInfo := createPaginationInfo(
		options.Limit,
		options.Offset,
		itemCount,
		false, // FilesResourceList doesn't provide total count
		0,
	)

	return &PagedFilesResourceList{
		FilesResourceList: files,
		Pagination:        paginationInfo,
	}, nil
}

// GetSortedFilesIterator returns an iterator for paginated access to sorted files
func (c *Client) GetSortedFilesIterator(options *PaginationOptions) *PaginationIterator[*PagedFilesResourceList] {
	fetcher := func(ctx context.Context, opts *PaginationOptions) (*PagedFilesResourceList, error) {
		result, errResp := c.GetSortedFilesPaged(ctx, opts)
		if errResp != nil {
			return nil, errors.New(errResp.Error)
		}
		return result, nil
	}

	return NewPaginationIterator(c, fetcher, options)
}

// GetLastUploadedResources returns the files uploaded most recently, newest
// first.
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {
	return c.GetLastUploadedResourcesWithPagination(ctx, nil)
}

// GetLastUploadedResourcesWithPagination gets last uploaded resources with pagination support
func (c *Client) GetLastUploadedResourcesWithPagination(ctx context.Context, options *PaginationOptions) (*LastUploadedResourceList, *ErrorResponse) {
	return c.GetLastUploadedResourcesWithOptions(ctx, options, nil)
}

// GetLastUploadedResourcesWithOptions is
// [Client.GetLastUploadedResourcesWithPagination] with filters: only some
// media types, and thumbnails of a chosen size. The API does not sort this
// list, so filters.Sort is ignored.
func (c *Client) GetLastUploadedResourcesWithOptions(ctx context.Context, options *PaginationOptions, filters *FilesOptions) (*LastUploadedResourceList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	query := url.Values{}
	addPaginationParams(query, options)
	filters.apply(query)
	query.Del("sort")
	query.Del("offset")

	endpoint := "resources/last-uploaded"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	return requestJSON[LastUploadedResourceList](ctx, c, GET, endpoint, nil)
}

// GetLastUploadedResourcesPaged returns a paginated wrapper with pagination info
func (c *Client) GetLastUploadedResourcesPaged(ctx context.Context, options *PaginationOptions) (*PagedLastUploadedResourceList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	files, errResp := c.GetLastUploadedResourcesWithPagination(ctx, options)
	if errResp != nil {
		return nil, errResp
	}

	// Create pagination info
	itemCount := len(files.Items)
	paginationInfo := createPaginationInfo(
		options.Limit,
		options.Offset,
		itemCount,
		false, // LastUploadedResourceList doesn't provide total count
		0,
	)

	return &PagedLastUploadedResourceList{
		LastUploadedResourceList: files,
		Pagination:               paginationInfo,
	}, nil
}

// GetLastUploadedResourcesIterator returns an iterator for paginated access to last uploaded resources
func (c *Client) GetLastUploadedResourcesIterator(options *PaginationOptions) *PaginationIterator[*PagedLastUploadedResourceList] {
	fetcher := func(ctx context.Context, opts *PaginationOptions) (*PagedLastUploadedResourceList, error) {
		result, errResp := c.GetLastUploadedResourcesPaged(ctx, opts)
		if errResp != nil {
			return nil, errors.New(errResp.Error)
		}
		return result, nil
	}

	return NewPaginationIterator(c, fetcher, options)
}

// MoveResource moves or renames the file or folder at from to path. It fails
// if path already exists. Moving a large folder may continue in the
// background; the returned [Link] then points to the operation, see
// [Client.GetOperationStatus].
func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	return c.MoveResourceWithOptions(ctx, from, path, nil)
}

// MoveResourceWithOptions is [Client.MoveResource] with the choice to
// overwrite the target or to run in the background.
func (c *Client) MoveResourceWithOptions(ctx context.Context, from, path string, opts *CopyMoveOptions) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)
	opts.apply(query)

	return requestJSON[Link](ctx, c, POST, "resources/move?"+query.Encode(), nil,
		http.StatusCreated, http.StatusAccepted)
}

// GetPublicResources returns the first page of the user's published files
// and folders.
func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	return c.GetPublicResourcesWithPagination(ctx, nil)
}

// GetPublicResourcesWithPagination gets public resources with pagination support
func (c *Client) GetPublicResourcesWithPagination(ctx context.Context, options *PaginationOptions) (*PublicResourcesList, *ErrorResponse) {
	return c.GetPublicResourcesWithOptions(ctx, options, nil)
}

// PublicResourcesOptions are the filters of
// [Client.GetPublicResourcesWithOptions].
type PublicResourcesOptions struct {
	// Type is "file" or "dir" to list only files or only folders; empty
	// lists both.
	Type        string
	PreviewSize string   // Thumbnail size, e.g. "M" or "120x240"
	PreviewCrop bool     // Crop previews to the requested size
	Fields      []string // Response fields to return (empty = all)
}

// GetPublicResourcesWithOptions is [Client.GetPublicResourcesWithPagination]
// with filters: only files or only folders, and thumbnails of a chosen size.
func (c *Client) GetPublicResourcesWithOptions(ctx context.Context, options *PaginationOptions, filters *PublicResourcesOptions) (*PublicResourcesList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	query := url.Values{}
	addPaginationParams(query, options)
	if filters != nil {
		if filters.Type != "" {
			query.Set("type", filters.Type)
		}
		if filters.PreviewSize != "" {
			query.Set("preview_size", filters.PreviewSize)
		}
		if filters.PreviewCrop {
			query.Set("preview_crop", "true")
		}
		if len(filters.Fields) > 0 {
			query.Set("fields", strings.Join(filters.Fields, ","))
		}
	}

	endpoint := "resources/public"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	return requestJSON[PublicResourcesList](ctx, c, GET, endpoint, nil)
}

// GetPublicResourcesPaged returns a paginated wrapper with pagination info
func (c *Client) GetPublicResourcesPaged(ctx context.Context, options *PaginationOptions) (*PagedPublicResourcesList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	list, errResp := c.GetPublicResourcesWithPagination(ctx, options)
	if errResp != nil {
		return nil, errResp
	}

	// Create pagination info - PublicResourcesList includes limit and offset
	itemCount := len(list.Items)
	paginationInfo := createPaginationInfo(
		list.Limit,  // Use actual limit from response
		list.Offset, // Use actual offset from response
		itemCount,
		false, // PublicResourcesList doesn't provide total count
		0,
	)

	return &PagedPublicResourcesList{
		PublicResourcesList: list,
		Pagination:          paginationInfo,
	}, nil
}

// GetPublicResourcesIterator returns an iterator for paginated access to public resources
func (c *Client) GetPublicResourcesIterator(options *PaginationOptions) *PaginationIterator[*PagedPublicResourcesList] {
	fetcher := func(ctx context.Context, opts *PaginationOptions) (*PagedPublicResourcesList, error) {
		result, errResp := c.GetPublicResourcesPaged(ctx, opts)
		if errResp != nil {
			return nil, errors.New(errResp.Error)
		}
		return result, nil
	}

	return NewPaginationIterator(c, fetcher, options)
}

// PublishResource publishes the file or folder at path, so anyone with the
// link can open it. The returned [Link] points to the resource; fetch it with
// [Client.GetMetadata] to read [Resource.PublicURL].
func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources/publish?"+query.Encode(), nil)
}

// UnpublishResource withdraws the public link of the file or folder at path.
func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources/unpublish?"+query.Encode(), nil)
}

// GetLinkForUpload returns where to send the contents of a new file at path.
// Most callers want [Client.UploadFileFromPath] instead, which does both
// steps.
func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[ResourceUploadLink](ctx, c, GET, "resources/upload?"+query.Encode(), nil)
}

// UploadFile asks Yandex.Disk to download the file at uploadURL from the
// internet and store it at path. The download happens on Yandex's side and
// in the background; the returned [Link] points to the operation, see
// [Client.GetOperationStatus].
func (c *Client) UploadFile(ctx context.Context, path, uploadURL string) (*Link, *ErrorResponse) {
	return c.UploadFileWithOptions(ctx, path, uploadURL, nil)
}

// UploadFromURLOptions are the optional parameters of
// [Client.UploadFileWithOptions].
type UploadFromURLOptions struct {
	// DisableRedirects makes Yandex refuse to follow redirects from
	// uploadURL.
	DisableRedirects bool
}

// UploadFileWithOptions is [Client.UploadFile] with the choice not to follow
// redirects.
func (c *Client) UploadFileWithOptions(ctx context.Context, path, uploadURL string, opts *UploadFromURLOptions) (*Link, *ErrorResponse) {
	if len(path) < 1 || len(uploadURL) < 1 {
		return nil, &ErrorResponse{Error: "path and url cannot be empty"}
	}

	queryParams := url.Values{}
	queryParams.Set("path", path)
	queryParams.Set("url", uploadURL)
	if opts != nil && opts.DisableRedirects {
		queryParams.Set("disable_redirects", "true")
	}

	return requestJSON[Link](ctx, c, POST, "resources/upload?"+queryParams.Encode(), nil,
		http.StatusOK, http.StatusAccepted)
}
