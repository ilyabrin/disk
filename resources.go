package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (c *Client) buildDeleteResourceURL(path string, permanently bool) string {
	query := url.Values{}
	query.Set("path", path)
	query.Set("permanent", strconv.FormatBool(permanently))
	return fmt.Sprintf("resources?%s", query.Encode())
}

// todo: add *ErrorResponse to return
func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	if path == "" {
		return errors.New("delete error: path cannot be empty")
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

func (c *Client) GetMetadata(ctx context.Context, path string) (*Resource, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var resource *Resource
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, GET, "resources?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode resource: %v", err)}
	}
	return resource, nil
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

	var resource *Resource
	var errorResponse *ErrorResponse

	body, err := json.Marshal(custom_properties)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to marshal properties: %v", err)}
	}

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, PATCH, "resources?"+query.Encode(), bytes.NewBuffer(body))
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode resource: %v", err)}
	}
	return resource, nil
}

// CreateDir creates a new directory with the specified 'path' name.
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, PUT, "resources?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}
	return link, nil
}

func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)
	resp, err := c.doRequest(ctx, POST, "resources/copy?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if !inArray(resp.StatusCode, []int{200, 201, 202}) {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}
	return link, nil
}

func (c *Client) GetDownloadURL(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, GET, "resources/download?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}
	return link, nil
}

func (c *Client) GetSortedFiles(ctx context.Context) (*FilesResourceList, *ErrorResponse) {
	return c.GetSortedFilesWithPagination(ctx, nil)
}

// GetSortedFilesWithPagination gets a sorted list of files with pagination support
func (c *Client) GetSortedFilesWithPagination(ctx context.Context, options *PaginationOptions) (*FilesResourceList, *ErrorResponse) {
	var files *FilesResourceList
	var errorResponse *ErrorResponse

	// Validate and normalize pagination options
	options = ValidatePaginationOptions(options)

	// Build query parameters
	query := url.Values{}
	addPaginationParams(query, options)

	endpoint := "resources/files"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, GET, endpoint, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&files); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode files: %v", err)}
	}
	return files, nil
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
			return nil, fmt.Errorf(errResp.Error)
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
	var files *LastUploadedResourceList
	var errorResponse *ErrorResponse

	// Validate and normalize pagination options
	options = ValidatePaginationOptions(options)

	// Build query parameters
	query := url.Values{}
	addPaginationParams(query, options)

	endpoint := "resources/last-uploaded"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, GET, endpoint, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&files); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode files: %v", err)}
	}

	return files, nil
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
			return nil, fmt.Errorf(errResp.Error)
		}
		return result, nil
	}

	return NewPaginationIterator(c, fetcher, options)
}

func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, &ErrorResponse{Error: "from and path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("from", from)
	query.Set("path", path)
	resp, err := c.doRequest(ctx, POST, "resources/move?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if !inArray(resp.StatusCode, []int{201, 202}) {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}

func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	return c.GetPublicResourcesWithPagination(ctx, nil)
}

// GetPublicResourcesWithPagination gets public resources with pagination support
func (c *Client) GetPublicResourcesWithPagination(ctx context.Context, options *PaginationOptions) (*PublicResourcesList, *ErrorResponse) {
	var list *PublicResourcesList
	var errorResponse *ErrorResponse

	// Validate and normalize pagination options
	options = ValidatePaginationOptions(options)

	// Build query parameters
	query := url.Values{}
	addPaginationParams(query, options)

	endpoint := "resources/public"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, GET, endpoint, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&list); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode list: %v", err)}
	}

	return list, nil
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
			return nil, fmt.Errorf(errResp.Error)
		}
		return result, nil
	}

	return NewPaginationIterator(c, fetcher, options)
}

func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, PUT, "resources/publish?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}

func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, PUT, "resources/unpublish?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	if len(path) < 1 {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	var resource *ResourceUploadLink
	var errorResponse *ErrorResponse

	query := url.Values{}
	query.Set("path", path)
	resp, err := c.doRequest(ctx, GET, "resources/upload?"+query.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode resource: %v", err)}
	}

	return resource, nil
}

// todo: empty resonses - fix it
func (c *Client) UploadFile(ctx context.Context, path, uploadURL string) (*Link, *ErrorResponse) {
	if len(path) < 1 || len(uploadURL) < 1 {
		return nil, &ErrorResponse{Error: "path and url cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	queryParams := url.Values{}
	queryParams.Set("path", path)
	queryParams.Set("url", uploadURL)
	resp, err := c.doRequest(ctx, POST, "resources/upload?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if !inArray(resp.StatusCode, []int{200, 202}) {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}
