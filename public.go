package disk

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string, params *QueryParams) (*PublicResource, *ErrorResponse) {
	var resource *PublicResource

	resp, err := c.get(ctx, c.apiURL+"public/resources?public_key="+public_key, params)
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

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.get(ctx, c.apiURL+"public/resources/download?public_key="+public_key, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := c.post(ctx, c.apiURL+"public/resources/save-to-disk?public_key="+public_key, nil, nil, params)
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
