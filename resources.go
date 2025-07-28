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
	var files *FilesResourceList
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, GET, "resources/files", nil)
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

// get | sortBy = [name = default, uploadDate]
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {
	var files *LastUploadedResourceList
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, GET, "resources/last-uploaded", nil)
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
	var list *PublicResourcesList
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, GET, "resources/public", nil)
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
