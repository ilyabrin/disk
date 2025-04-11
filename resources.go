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

// DeleteResource deletes a resource at the specified path.
func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	if path == "" {
		return errors.New("delete error: empty path")
	}

	_, err := doRequest[struct{}](ctx, c, DELETE, c.buildDeleteResourceURL(path, permanently), nil)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata for a resource at the specified path.
func (c *Client) GetMetadata(ctx context.Context, path string) (*Resource, *ErrorResponse) {
	if path == "" {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	resource, err := doRequest[*Resource](ctx, c, GET, "resources?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return resource, nil
}

// UpdateMetadata updates custom properties for a resource at the specified path.
func (c *Client) UpdateMetadata(ctx context.Context, path string, customProperties map[string]map[string]string) (*Resource, *ErrorResponse) {
	if path == "" {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	body, err := json.Marshal(customProperties)
	if err != nil {
		return nil, &ErrorResponse{Error: "failed to marshal custom properties"}
	}

	resource, err := doRequest[*Resource](ctx, c, PATCH, "resources?path="+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return resource, nil
}

// CreateDir creates a new directory with the specified 'path' name.
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if path == "" {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	link, err := doRequest[*Link](ctx, c, PUT, "resources?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// CopyResource copies a resource from one path to another.
func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if from == "" || path == "" {
		return nil, &ErrorResponse{Error: "source and destination paths cannot be empty"}
	}

	link, err := doRequest[*Link](ctx, c, POST, "resources/copy?from="+from+"&path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// GetDownloadURL retrieves the download URL for a resource at the specified path.
func (c *Client) GetDownloadURL(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if path == "" {
		return nil, &ErrorResponse{Error: "path cannot be empty"}
	}

	link, err := doRequest[*Link](ctx, c, GET, "resources/download?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// GetSortedFiles retrieves a sorted list of files.
func (c *Client) GetSortedFiles(ctx context.Context) (*FilesResourceList, *ErrorResponse) {
	files, err := doRequest[*FilesResourceList](ctx, c, GET, "resources/files", nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return files, nil
}

// GetLastUploadedResources retrieves the last uploaded resources.
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {
	files, err := doRequest[*LastUploadedResourceList](ctx, c, GET, "resources/last-uploaded", nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return files, nil
}

// MoveResource moves a resource from one path to another.
func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, POST, "resources/move?from="+from+"&path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// GetPublicResources retrieves a list of public resources.
func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	list, err := doRequest[*PublicResourcesList](ctx, c, GET, "resources/public", nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return list, nil
}

// PublishResource publishes a resource at the specified path.
func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, PUT, "resources/publish?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// UnpublishResource unpublishes a resource at the specified path.
func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, PUT, "resources/unpublish?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// GetLinkForUpload retrieves a link for uploading a resource at the specified path.
func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	resource, err := doRequest[*ResourceUploadLink](ctx, c, GET, "resources/upload?path="+path, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return resource, nil
}

// UploadFile uploads a file to the specified path using the provided URL.
func (c *Client) UploadFile(ctx context.Context, path, url string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, POST, "resources/upload?path="+path+"&url="+url, nil)
	if err != nil {
		return nil, c.parseErrorResponse(err)
	}

	return link, nil
}

// parseErrorResponse parses an error into an ErrorResponse if possible.
func (c *Client) parseErrorResponse(err error) *ErrorResponse {
	var errorResponse *ErrorResponse
	if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
		return errorResponse
	}
	return &ErrorResponse{Error: "unknown error occurred"}
}
