package disk

import (
	"context"
)

func (c *Client) DiskInfo(ctx context.Context) (*Disk, error) {
	// Use the generic doRequest function to simplify the code
	disk, err := doRequest[*Disk](ctx, c, GET, "", nil)
	if err != nil {
		return nil, err
	}

	return disk, nil
}
