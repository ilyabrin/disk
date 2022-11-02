package disk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
)

// todo: add *ErrorResponse to return

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

	newMeta := map[string]map[string]string{
		"custom_properties": {
			"key_01": "value_01",
			"key_02": "value_02",
			"key_07": "value_07",
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

	// if  {
	// 	decoded = json.NewDecoder(resp.Body)
	// 	err := decoded.Decode(&errorResponse)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return nil, nil
	// 	}
	// 	return nil, errorResponse
	// }

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return link, nil
}

func (c *Client) CopyResource(ctx context.Context, from, path string, params *queryParams) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, nil
	}

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "POST", c.api_url+"resources/copy?from="+from+"&path="+path, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if !inArray(resp.StatusCode, []int{200, 201, 202}) {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if err != nil {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return link, nil
}

func (c *Client) GetDownloadURL(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"resources/download?path="+path, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return link, nil
}

func (c *Client) GetSortedFiles(ctx context.Context, params *queryParams) (*FilesResourceList, *ErrorResponse) {

	var files *FilesResourceList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"resources/files", nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&files); err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return files, nil
}

// get | sortBy = [name = default, uploadDate]
func (c *Client) GetLastUploadedResources(ctx context.Context, params *queryParams) (*LastUploadedResourceList, *ErrorResponse) {

	var files *LastUploadedResourceList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"resources/last-uploaded", nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&files); err != nil {
		log.Fatal(err)
		return nil, nil
	}

	return files, nil
}

func (c *Client) MoveResource(ctx context.Context, from, path string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "POST", c.api_url+"resources/move?from="+from+"&path="+path, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if !inArray(resp.StatusCode, []int{201, 202}) {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil
}

func (c *Client) GetPublicResources(ctx context.Context, params *queryParams) (*PublicResourcesList, *ErrorResponse) {
	var list *PublicResourcesList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"resources/public", nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&list); err != nil {
		log.Fatal(err)
	}

	return list, nil
}

func (c *Client) PublishResource(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "PUT", c.api_url+"resources/publish?path="+path, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil
}

func (c *Client) UnpublishResource(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "PUT", c.api_url+"resources/unpublish?path="+path, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil
}

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*Link, *ErrorResponse) {
	var resource *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"resources/upload?path="+path, nil, nil, nil)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		log.Fatal(err)
	}

	return resource, nil
}

func (c *Client) UploadFile(ctx context.Context, file_path, url string, params *queryParams) *ErrorResponse {

	var decoded *json.Decoder
	var errorResponse *ErrorResponse

	f, err := os.Open(file_path)
	if err != nil {
		log.Fatal(err)
	}
	body := bufio.NewReader(f)
	defer f.Close()

	resp, err := c.doRequest(ctx, "PUT", url, body, nil, nil)
	if haveError(err) {
		decoded = json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			log.Fatal(err)
		}
	}

	return handleResponseCode(resp.StatusCode)
}
