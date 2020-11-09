package disk

import (
	"context"
	"encoding/json"
	"log"
)

func (c *Client) DiskInfo(ctx context.Context) (*Disk, error) {
	var disk *Disk
	resp, _ := c.doRequest(ctx, GET, "disk", nil)

	decoded := json.NewDecoder(resp.Body)
	// decoded.DisallowUnknownFields()

	if err := decoded.Decode(&disk); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return disk, nil
}
