package disk

import (
	"context"
)

func (c *Client) DeleteFromTrash(ctx context.Context, path string) {
	// TODO
}

func (c *Client) DeleteAllFromTrash(ctx context.Context) {
	c.DeleteFromTrash(ctx, "trash root here")
}

func (c *Client) RestoreFromTrash(ctx context.Context, path string) {
	// TODO
}
