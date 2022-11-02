package disk

import (
	"context"
	"encoding/json"
)

func (c *Client) DeleteFromTrash(ctx context.Context, path string, params *queryParams) (*Link, *ErrorResponse) {

	resp, err := c.delete(ctx, c.api_url+"trash/resources?path="+path, nil, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}

	var link *Link
	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
			return nil, jsonDecodeError(err)
		}
	}

	return nil, nil
}

func (c *Client) RestoreFromTrash(ctx context.Context, path string, params *queryParams) (*Link, *Operation, *ErrorResponse) {

	var link *Link

	resp, err := c.put(ctx, c.api_url+"trash/resources/restore?path="+path, nil, nil, params)
	if haveError(err) {
		return nil, nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
		return nil, nil, jsonDecodeError(err)
	}

	return link, nil, nil
}

func (c *Client) ListTrashResources(ctx context.Context, path string, params *queryParams) (*TrashResource, *ErrorResponse) {

	var resource *TrashResource

	resp, err := c.get(ctx, c.api_url+"trash/resources?path="+path, nil, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}
