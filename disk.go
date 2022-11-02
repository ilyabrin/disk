package disk

import (
	"context"
	"encoding/json"
)

// TODO: add APIResponse
func (c *Client) DiskInfo(ctx context.Context, params *QueryParams) (*Disk, *ErrorResponse) {
	var disk *Disk

	resp, err := c.get(ctx, c.apiURL, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&disk)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	return disk, nil
}
