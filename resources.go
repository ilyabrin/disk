package disk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
)

func (c *Client) DeleteResource(ctx context.Context, path string, permanent bool, params *queryParams) *ErrorResponse {

	url := "resources?path=" + path + "&permanent=false"
	if permanent {
		url = "resources?path=" + path + "&permanent=true"
	}

	resp, err := c.delete(ctx, c.api_url+url, nil, params)
	if haveError(err) {
		return handleResponseCode(resp.StatusCode)
	}

	return nil
}

func (c *Client) GetMetadata(ctx context.Context, path string, params *queryParams) (*Resource, *ErrorResponse) {

	var resource *Resource

	resp, err := c.get(ctx, c.api_url+"resources?path="+path, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

/*
todo: add examples to README

	newMeta := &metadata{
		"custom_properties": {
			"key": "value",
			"foo": "bar",
			"platform": "linux",
		},
	}
*/
func (c *Client) UpdateMetadata(ctx context.Context, path string, custom_properties *metadata) (*Resource, *ErrorResponse) {

	var resource *Resource
	var body []byte

	body, err := json.Marshal(custom_properties)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	resp, err := c.patch(ctx, c.api_url+"resources?path="+path, bytes.NewBuffer([]byte(body)), nil, nil)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

// CreateDir creates a new dorectory with 'path'(string) name
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.put(ctx, c.api_url+"resources?path="+path, nil, nil, params)

	if haveError(err) || resp.StatusCode != 201 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) CopyResource(ctx context.Context, from, to string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.post(ctx, c.api_url+"resources/copy?from="+from+"&path="+to, nil, nil, params)
	if haveError(err) || !inArray(resp.StatusCode, []int{200, 201, 202}) {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetDownloadURL(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.get(ctx, c.api_url+"resources/download?path="+path, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetSortedFiles(ctx context.Context, params *queryParams) (*FilesResourceList, *ErrorResponse) {

	var files *FilesResourceList

	resp, err := c.get(ctx, c.api_url+"resources/files", nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

// get | sortBy = [name = default, uploadDate]
func (c *Client) GetLastUploadedResources(ctx context.Context, params *queryParams) (*LastUploadedResourceList, *ErrorResponse) {

	var files *LastUploadedResourceList

	resp, err := c.get(ctx, c.api_url+"resources/last-uploaded", nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

func (c *Client) MoveResource(ctx context.Context, from, to string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.post(ctx, c.api_url+"resources/move?from="+from+"&path="+to, nil, nil, params)
	if haveError(err) || !inArray(resp.StatusCode, []int{201, 202}) {

	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetPublicResources(ctx context.Context, params *queryParams) (*PublicResourcesList, *ErrorResponse) {

	var list *PublicResourcesList

	resp, err := c.get(ctx, c.api_url+"resources/public", nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, jsonDecodeError(err)
	}

	return list, nil
}

func (c *Client) PublishResource(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.put(ctx, c.api_url+"resources/publish?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) UnpublishResource(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.put(ctx, c.api_url+"resources/unpublish?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*Link, *ErrorResponse) {

	var resource *Link

	resp, err := c.get(ctx, c.api_url+"resources/upload?path="+path, nil, nil)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

func (c *Client) UploadFile(ctx context.Context, file_path, url string, params *queryParams) *ErrorResponse {

	var errorResponse *ErrorResponse

	f, err := os.Open(file_path)
	if err != nil {
		return jsonDecodeError(err)
	}
	body := bufio.NewReader(f)
	defer f.Close()

	resp, err := c.put(ctx, url, body, nil, nil)
	if haveError(err) {
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return jsonDecodeError(err)
		}
	}

	return handleResponseCode(resp.StatusCode)
}
