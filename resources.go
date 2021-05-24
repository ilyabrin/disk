package disk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// todo: add *ErrorResponse to return
func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	if len(path) < 1 {
		return errors.New("delete error")
	}

	var url string

	// todo: make it better
	if permanently {
		url = "resources?path=" + path + "&permanent=true"
	} else {
		url = "resources?path=" + path + "&permanent=false"
	}

	resp, err := c.doRequest(ctx, DELETE, url, nil)
	if haveError(err) {
		log.Fatal(err)
		return err
	}

	fmt.Println(resp.Body)

	return nil
}

func (c *Client) GetMetadata(ctx context.Context, path string) (*Resource, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	var resource *Resource
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources?path="+path, nil)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if err != nil {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return resource, nil
}

/* todo: add examples to README
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
		return nil, nil
	}

	var resource *Resource
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	var body []byte

	body, err = json.Marshal(custom_properties)

	if haveError(err) {
		log.Fatal("payload error")
	}

	resp, err := c.doRequest(ctx, PATCH, "resources?path="+path, bytes.NewBuffer([]byte(body)))
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if err != nil {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		log.Fatal(err)
		return nil, nil
	}
	return resource, nil
}

// CreateDir creates a new dorectory with 'path'(string) name
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, PUT, "resources?path="+path, nil)
	if haveError(err) {
		log.Fatal("Request failed")
		return nil, nil
	}

	if resp.StatusCode != 201 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if err != nil {
			log.Fatal(err)
			return nil, nil
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

func (c *Client) CopyResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {
	if len(from) < 1 || len(path) < 1 {
		return nil, nil
	}

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, POST, "resources/copy?from="+from+"&path="+path, nil)
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

func (c *Client) GetDownloadURL(ctx context.Context, path string) (*Link, *ErrorResponse) {
	if len(path) < 1 {
		return nil, nil
	}

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources/download?path="+path, nil)
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

func (c *Client) GetSortedFiles(ctx context.Context) (*FilesResourceList, *ErrorResponse) {

	var files *FilesResourceList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources/files", nil)
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
func (c *Client) GetLastUploadedResources(ctx context.Context) (*LastUploadedResourceList, *ErrorResponse) {

	var files *LastUploadedResourceList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources/last-uploaded", nil)
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

func (c *Client) MoveResource(ctx context.Context, from, path string) (*Link, *ErrorResponse) {

	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, POST, "resources/move?from="+from+"&path="+path, nil)
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

func (c *Client) GetPublicResources(ctx context.Context) (*PublicResourcesList, *ErrorResponse) {
	var list *PublicResourcesList
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources/public", nil)
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

func (c *Client) PublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, PUT, "resources/publish?path="+path, nil)
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

func (c *Client) UnpublishResource(ctx context.Context, path string) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, PUT, "resources/unpublish?path="+path, nil)
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

func (c *Client) GetLinkForUpload(ctx context.Context, path string) (*ResourceUploadLink, *ErrorResponse) {
	var resource *ResourceUploadLink
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "resources/upload?path="+path, nil)
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

// todo: empty resonses - fix it
func (c *Client) UploadFile(ctx context.Context, path, url string) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, POST, "resources/upload?path="+path+"&url="+url, nil)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if !inArray(resp.StatusCode, []int{200, 202}) {
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
