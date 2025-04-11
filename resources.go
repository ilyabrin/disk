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
		// Handle cases where the response body is empty (e.g., HTTP 204)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("delete failed: %w", err)
		}
	}

	// Response was successful, return nil
	return nil
}

// GetMetadata retrieves metadata for a resource at the specified path.
func (c *Client) GetMetadata(ctx context.Context, path string) (*Resource, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	resource, err := doRequest[*Resource](ctx, c, GET, "resources?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return resource, nil
}

// UpdateMetadata updates custom properties for a resource at the specified path.
func (c *Client) UpdateMetadata(ctx context.Context, path string, custom_properties map[string]map[string]string) (*Resource, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	body, err := json.Marshal(custom_properties)
	if err != nil {
		return nil, nil
	}

	resource, err := doRequest[*Resource](ctx, c, PATCH, "resources?path="+path, bytes.NewBuffer(body))
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return resource, nil
}

// CreateDir creates a new directory with the specified 'path' name.
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	link, err := doRequest[*Link](ctx, c, PUT, "resources?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// CopyResource copies a resource from one path to another.
func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, nil
	}

	link, err := doRequest[*Link](ctx, c, POST, "resources/copy?from="+from+"&path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// GetDownloadURL retrieves the download URL for a resource at the specified path.
func (c *Client) GetDownloadURL(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	link, err := doRequest[*Link](ctx, c, GET, "resources/download?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// GetSortedFiles retrieves a sorted list of files.
func (c *Client) GetSortedFiles(ctx context.Context) (*FilesResourceList, *ErrorResponse) {
	files, err := doRequest[*FilesResourceList](ctx, c, GET, "resources/files", nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return files, nil
}

// GetLastUploadedResources retrieves the last uploaded resources.
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {
	files, err := doRequest[*LastUploadedResourceList](ctx, c, GET, "resources/last-uploaded", nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return files, nil
}

// MoveResource moves a resource from one path to another.
func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, POST, "resources/move?from="+from+"&path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// GetPublicResources retrieves a list of public resources.
func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	list, err := doRequest[*PublicResourcesList](ctx, c, GET, "resources/public", nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return list, nil
}

// PublishResource publishes a resource at the specified path.
func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, PUT, "resources/publish?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// UnpublishResource unpublishes a resource at the specified path.
func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, PUT, "resources/unpublish?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

// GetLinkForUpload retrieves a link for uploading a resource at the specified path.
func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	resource, err := doRequest[*ResourceUploadLink](ctx, c, GET, "resources/upload?path="+path, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return resource, nil
}

// UploadFile uploads a file to the specified path using the provided URL.
func (c *Client) UploadFile(ctx context.Context, path, url string) (*Link, *ErrorResponse) {
	link, err := doRequest[*Link](ctx, c, POST, "resources/upload?path="+path+"&url="+url, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}
