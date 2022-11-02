package disk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func (c *Client) DeleteResource(ctx context.Context, path string, permanent bool, params *QueryParams) *ErrorResponse {
	url := "resources?path=" + path + "&permanent=false"
	if permanent {
		url = "resources?path=" + path + "&permanent=true"
	}

	resp, err := c.delete(ctx, c.apiURL+url, nil, params)
	if haveError(err) {
		return handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) GetMetadata(ctx context.Context, path string, params *QueryParams) (*Resource, *ErrorResponse) {
	var resource *Resource

	resp, err := c.get(ctx, c.apiURL+"resources?path="+path, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

/*
todo: add examples to README

	newMeta := &disk.Metadata{
		"custom_properties": {
			"key": "value",
			"foo": "bar",
			"platform": "linux",
		},
	}
*/
func (c *Client) UpdateMetadata(ctx context.Context, path string, custom_properties *Metadata) (*Resource, *ErrorResponse) {
	var resource *Resource
	var body []byte

	body, err := json.Marshal(custom_properties)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	resp, err := c.patch(ctx, c.apiURL+"resources?path="+path, bytes.NewBuffer(body), nil, nil)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

// CreateDir creates a new dorectory with 'path'(string) name
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.put(ctx, c.apiURL+"resources?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != http.StatusCreated {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) CopyResource(ctx context.Context, from, to string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.post(ctx, c.apiURL+"resources/copy?from="+from+"&path="+to, nil, nil, params)
	if haveError(err) || !InArray(resp.StatusCode, []int{200, 201, 202}) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetDownloadURL(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.get(ctx, c.apiURL+"resources/download?path="+path, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetSortedFiles(ctx context.Context, params *QueryParams) (*FilesResourceList, *ErrorResponse) {
	var files *FilesResourceList

	resp, err := c.get(ctx, c.apiURL+"resources/files", params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

// get | sortBy = [name = default, uploadDate]
func (c *Client) GetLastUploadedResources(ctx context.Context, params *QueryParams) (*LastUploadedResourceList, *ErrorResponse) {
	var files *LastUploadedResourceList

	resp, err := c.get(ctx, c.apiURL+"resources/last-uploaded", params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

func (c *Client) MoveResource(ctx context.Context, from, to string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.post(ctx, c.apiURL+"resources/move?from="+from+"&path="+to, nil, nil, params)
	if haveError(err) || !InArray(resp.StatusCode, []int{201, 202}) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetPublicResources(ctx context.Context, params *QueryParams) (*PublicResourcesList, *ErrorResponse) {
	var list *PublicResourcesList

	resp, err := c.get(ctx, c.apiURL+"resources/public", params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return list, nil
}

func (c *Client) PublishResource(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.put(ctx, c.apiURL+"resources/publish?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) UnpublishResource(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.put(ctx, c.apiURL+"resources/unpublish?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*Link, *ErrorResponse) {
	var resource *Link

	resp, err := c.get(ctx, c.apiURL+"resources/upload?path="+path, nil)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

func (c *Client) UploadFile(ctx context.Context, file, url string, params *QueryParams) *ErrorResponse {
	var errorResponse *ErrorResponse

	f, err := os.Open(file)
	if err != nil {
		return jsonDecodeError(err)
	}
	body := bufio.NewReader(f)
	defer f.Close()

	headers := &HTTPHeaders{
		"Content-Type": "multipart/form-data",
	}
	resp, err := c.put(ctx, url, body, headers, nil)
	if haveError(err) {
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return jsonDecodeError(err)
		}
	}
	defer resp.Body.Close()

	return handleResponseCode(resp.StatusCode)
}
