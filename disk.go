package disk

import (
	"context"
	"encoding/json"
)

// TODO: add APIResponse
func (c *Client) DiskInfo(ctx context.Context, params *queryParams) (*Disk, *ErrorResponse) {

	var disk *Disk

	resp, err := c.get(ctx, c.api_url, nil, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&disk); err != nil {
		return nil, jsonDecodeError(err)
	}

	return disk, nil
}
