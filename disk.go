package disk

import (
	"context"
	"fmt"
)

func (c *Client) DiskInfo(ctx context.Context) (*Disk, error) {
	var disk *Disk
	resp, err := c.doRequest(ctx, GET, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}
	defer resp.Body.Close()

	// Use centralized response handling
	if _, err := c.handleResponse(resp, []int{200}); err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	// Use safe JSON decoding
	if err := c.safeDecodeJSON(resp, &disk); err != nil {
		return nil, fmt.Errorf("failed to decode disk info: %w", err)
	}

	return disk, nil
}
