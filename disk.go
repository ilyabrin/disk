package disk

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) DiskInfo(ctx context.Context) (*Disk, error) {
	var disk *Disk
	resp, err := c.doRequest(ctx, GET, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errorResponse ErrorResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorResponse); decodeErr != nil {
			return nil, fmt.Errorf("request failed with status %d: %w", resp.StatusCode, decodeErr)
		}
		return nil, fmt.Errorf("request failed: %s", errorResponse.Error)
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&disk); err != nil {
		return nil, fmt.Errorf("failed to decode disk info: %w", err)
	}

	return disk, nil
}
