package disk

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// RestoreFromTrash restores a resource from trash to its original location or a new path
func (c *Client) RestoreFromTrash(ctx context.Context, path string, overwrite bool, name string) (*Link, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	query := url.Values{}
	query.Set("path", path)
	if overwrite {
		query.Set("overwrite", "true")
	}
	if name != "" {
		query.Set("name", name)
	}

	c.Logger.Debug("Restoring resource from trash: %s", path)

	resp, err := c.doRequest(ctx, PUT, "trash/resources/restore?"+query.Encode(), nil)
	if err != nil {
		c.Logger.LogError("restore from trash", err)
		return nil, fmt.Errorf("failed to restore from trash: %w", err)
	}
	defer resp.Body.Close()

	// Handle different response codes
	if _, err := c.handleResponse(resp, []int{200, 201, 202}); err != nil {
		return nil, fmt.Errorf("failed to restore from trash: %w", err)
	}

	// 201 carries a link to the restored resource, 202 a link to the
	// asynchronous operation. Both are worth returning to the caller.
	var link Link
	if resp.StatusCode != 204 {
		if err := c.safeDecodeJSON(resp, &link); err != nil {
			return nil, fmt.Errorf("failed to decode restore response: %w", err)
		}
	}

	c.Logger.Info("Successfully restored resource from trash: %s", path)
	return &link, nil
}

// ListTrashResources lists resources in the trash, optionally filtered by path
func (c *Client) ListTrashResources(ctx context.Context, path string, limit int, offset int) (*TrashResourceList, error) {
	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	c.Logger.Debug("Listing trash resources with path: %s", path)

	resp, err := c.doRequest(ctx, GET, "trash/resources?"+query.Encode(), nil)
	if err != nil {
		c.Logger.LogError("list trash resources", err)
		return nil, fmt.Errorf("failed to list trash resources: %w", err)
	}
	defer resp.Body.Close()

	if _, err := c.handleResponse(resp, []int{200}); err != nil {
		return nil, fmt.Errorf("failed to list trash resources: %w", err)
	}

	var trashList TrashResourceList
	if err := c.safeDecodeJSON(resp, &trashList); err != nil {
		return nil, fmt.Errorf("failed to decode trash list: %w", err)
	}

	c.Logger.Info("Successfully listed %d trash resources", len(trashList.Items))
	return &trashList, nil
}

// EmptyTrash permanently deletes all resources from trash or a specific path in trash.
// Set forceAsync to make the API always run the deletion in the background and
// answer 202 with a link to the operation.
func (c *Client) EmptyTrash(ctx context.Context, path string, forceAsync bool) error {
	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}
	if forceAsync {
		query.Set("force_async", "true")
	}

	c.Logger.Debug("Emptying trash with path: %s", path)

	resp, err := c.doRequest(ctx, DELETE, "trash/resources?"+query.Encode(), nil)
	if err != nil {
		c.Logger.LogError("empty trash", err)
		return fmt.Errorf("failed to empty trash: %w", err)
	}
	defer resp.Body.Close()

	// Handle different response codes
	if _, err := c.handleResponse(resp, []int{200, 202, 204}); err != nil {
		return fmt.Errorf("failed to empty trash: %w", err)
	}

	if resp.StatusCode == 202 {
		c.Logger.Info("Trash emptying started asynchronously")
	} else {
		c.Logger.Info("Successfully emptied trash")
	}

	return nil
}

// GetTrashResourceMetadata retrieves metadata for a specific resource in trash
func (c *Client) GetTrashResourceMetadata(ctx context.Context, path string, fields []string) (*TrashResource, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	query := url.Values{}
	query.Set("path", path)
	if len(fields) > 0 {
		for _, field := range fields {
			query.Add("fields", field)
		}
	}

	c.Logger.Debug("Getting trash resource metadata: %s", path)

	resp, err := c.doRequest(ctx, GET, "trash/resources?"+query.Encode(), nil)
	if err != nil {
		c.Logger.LogError("get trash resource metadata", err)
		return nil, fmt.Errorf("failed to get trash resource metadata: %w", err)
	}
	defer resp.Body.Close()

	if _, err := c.handleResponse(resp, []int{200}); err != nil {
		return nil, fmt.Errorf("failed to get trash resource metadata: %w", err)
	}

	var trashResource TrashResource
	if err := c.safeDecodeJSON(resp, &trashResource); err != nil {
		return nil, fmt.Errorf("failed to decode trash resource metadata: %w", err)
	}

	c.Logger.Info("Successfully retrieved trash resource metadata: %s", path)
	return &trashResource, nil
}
