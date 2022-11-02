package disk

import (
	"context"
	"encoding/json"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string, params *queryParams) (*PublicResource, *ErrorResponse) {

	var resource *PublicResource

	resp, err := c.get(ctx, c.api_url+"public/resources?public_key="+public_key, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.get(ctx, c.api_url+"public/resources/download?public_key="+public_key, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string, params *queryParams) (*Link, *ErrorResponse) {

	var link *Link

	resp, err := c.post(ctx, c.api_url+"public/resources/save-to-disk?public_key="+public_key, nil, nil, params)
	if haveError(err) || !inArray(resp.StatusCode, []int{200, 201, 202}) {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}
