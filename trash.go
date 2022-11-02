package disk

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) DeleteFromTrash(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	resp, err := c.delete(ctx, c.apiURL+"trash/resources?path="+path, nil, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	var link *Link

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&link)
		if err != nil {
			return nil, jsonDecodeError(err)
		}
	}

	return nil, nil
}

// RestoreFromTrash -
func (c *Client) RestoreFromTrash(ctx context.Context, path string, params *QueryParams) (*Link, *Operation, *ErrorResponse) {
	var link *Link

	resp, err := c.put(ctx, c.apiURL+"trash/resources/restore?path="+path, nil, nil, params)
	if haveError(err) {
		return nil, nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if haveError(err) {
		return nil, nil, jsonDecodeError(err)
	}

	return link, nil, nil
}

// ListTrashResources -
func (c *Client) ListTrashResources(ctx context.Context, path string, params *QueryParams) (*TrashResource, *ErrorResponse) {
	var resource *TrashResource

	resp, err := c.get(ctx, c.apiURL+"trash/resources?path="+path, params)
	if haveError(err) || resp.StatusCode != 200 {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}
