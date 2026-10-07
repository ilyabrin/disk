package disk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RestoreFromTrash restores a resource from trash to its original location or a new path
func (c *Client) RestoreFromTrash(ctx context.Context, path string, overwrite bool, name string) (*Link, error) {
	return c.RestoreFromTrashWithOptions(ctx, path, &RestoreOptions{Overwrite: overwrite, Name: name})
}

// RestoreOptions are the optional parameters of
// [Client.RestoreFromTrashWithOptions].
type RestoreOptions struct {
	// Overwrite replaces whatever is now at the original path.
	Overwrite bool
	// Name restores the resource under another name.
	Name string
	// ForceAsync makes the API restore in the background even when the
	// resource is small.
	ForceAsync bool
}

// RestoreFromTrashWithOptions is [Client.RestoreFromTrash] with the choice to
// run in the background.
func (c *Client) RestoreFromTrashWithOptions(ctx context.Context, path string, opts *RestoreOptions) (*Link, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if opts == nil {
		opts = &RestoreOptions{}
	}

	query := url.Values{}
	query.Set("path", path)
	if opts.Overwrite {
		query.Set("overwrite", "true")
	}
	if opts.Name != "" {
		query.Set("name", opts.Name)
	}
	if opts.ForceAsync {
		query.Set("force_async", "true")
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
	if resp.StatusCode != http.StatusNoContent {
		if err := c.safeDecodeJSON(resp, &link); err != nil {
			return nil, fmt.Errorf("failed to decode restore response: %w", err)
		}
	}

	c.Logger.Info("Successfully restored resource from trash: %s", path)
	return &link, nil
}

// ListTrashResources lists resources in the trash, optionally filtered by path
func (c *Client) ListTrashResources(ctx context.Context, path string, limit int, offset int) (*TrashResourceList, error) {
	return c.ListTrashResourcesWithOptions(ctx, path, &TrashListOptions{Limit: limit, Offset: offset})
}

// TrashListOptions are the optional parameters of
// [Client.ListTrashResourcesWithOptions].
type TrashListOptions struct {
	Limit       int      // Page size (0 = API default)
	Offset      int      // Items to skip
	Sort        string   // "deleted" or "created" (prefix "-" to reverse)
	PreviewSize string   // Thumbnail size, e.g. "M" or "120x240"
	PreviewCrop bool     // Crop previews to the requested size
	Fields      []string // Response fields to return (empty = all)
}

// ListTrashResourcesWithOptions is [Client.ListTrashResources] with sorting
// and thumbnails.
func (c *Client) ListTrashResourcesWithOptions(ctx context.Context, path string, opts *TrashListOptions) (*TrashResourceList, error) {
	if opts == nil {
		opts = &TrashListOptions{}
	}
	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.PreviewSize != "" {
		query.Set("preview_size", opts.PreviewSize)
	}
	if opts.PreviewCrop {
		query.Set("preview_crop", "true")
	}
	if len(opts.Fields) > 0 {
		query.Set("fields", strings.Join(opts.Fields, ","))
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

	// Yandex answers with the trash root as a resource, its contents under
	// "_embedded", exactly as for an ordinary folder. A list at the top level
	// is still accepted, for servers and fixtures that send that shape.
	var body struct {
		TrashResourceList
		Embedded *TrashResourceList `json:"_embedded"`
	}
	if err := c.safeDecodeJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("failed to decode trash list: %w", err)
	}
	trashList := body.TrashResourceList
	if body.Embedded != nil {
		trashList = *body.Embedded
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

	if resp.StatusCode == http.StatusAccepted {
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
