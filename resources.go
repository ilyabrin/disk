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
	query := url.Values{}
	query.Set("path", path)
	query.Set("permanent", strconv.FormatBool(permanently))
	return fmt.Sprintf("resources?%s", query.Encode())
}

// todo: add *ErrorResponse to return
func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	if err := validatePath(path); err != nil {
		return fmt.Errorf("delete error: %w", err)
	}

	url := c.buildDeleteResourceURL(path, permanently)

	resp, err := c.doRequest(ctx, DELETE, url, nil)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// Use centralized response handling
	if _, err := c.handleResponse(resp, []int{200}); err != nil {
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
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources?"+query.Encode(), nil, http.StatusCreated)
}

func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)

	return requestJSON[Link](ctx, c, POST, "resources/copy?"+query.Encode(), nil,
		http.StatusOK, http.StatusCreated, http.StatusAccepted)
}

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

// get | sortBy = [name = default, uploadDate]
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {
	return c.GetLastUploadedResourcesWithPagination(ctx, nil)
}

// GetLastUploadedResourcesWithPagination gets last uploaded resources with pagination support
func (c *Client) GetLastUploadedResourcesWithPagination(ctx context.Context, options *PaginationOptions) (*LastUploadedResourceList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	query := url.Values{}
	addPaginationParams(query, options)

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

func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)

	return requestJSON[Link](ctx, c, POST, "resources/move?"+query.Encode(), nil,
		http.StatusCreated, http.StatusAccepted)
}

func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	return c.GetPublicResourcesWithPagination(ctx, nil)
}

// GetPublicResourcesWithPagination gets public resources with pagination support
func (c *Client) GetPublicResourcesWithPagination(ctx context.Context, options *PaginationOptions) (*PublicResourcesList, *ErrorResponse) {
	options = ValidatePaginationOptions(options)

	query := url.Values{}
	addPaginationParams(query, options)

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

func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources/publish?"+query.Encode(), nil)
}

func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[Link](ctx, c, PUT, "resources/unpublish?"+query.Encode(), nil)
}

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	query := url.Values{}
	query.Set("path", path)

	return requestJSON[ResourceUploadLink](ctx, c, GET, "resources/upload?"+query.Encode(), nil)
}

// UploadFile asks Yandex Disk to fetch the file at uploadURL and store it at path.
// todo: empty resonses - fix it
func (c *Client) UploadFile(ctx context.Context, path, uploadURL string) (*Link, *ErrorResponse) {
	if len(path) < 1 || len(uploadURL) < 1 {
		return nil, &ErrorResponse{Error: "path and url cannot be empty"}
	}

	queryParams := url.Values{}
	queryParams.Set("path", path)
	queryParams.Set("url", uploadURL)

	return requestJSON[Link](ctx, c, POST, "resources/upload?"+queryParams.Encode(), nil,
		http.StatusOK, http.StatusAccepted)
}
