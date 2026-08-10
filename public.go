package disk

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PublicResourceOptions contains the optional parameters accepted by the
// /v1/disk/public/resources endpoints.
type PublicResourceOptions struct {
	Path        string   // Path to a resource inside a published folder
	Sort        string   // "name", "path", "created", "modified", "size" (prefix "-" to reverse)
	Limit       int      // Max number of nested items to return (0 = API default)
	Offset      int      // Number of nested items to skip
	PreviewSize string   // Thumbnail size, e.g. "M" or "120x240"
	PreviewCrop bool     // Crop previews to the requested size
	Fields      []string // Response fields to return (empty = all)
}

func (o *PublicResourceOptions) apply(query url.Values) {
	if o == nil {
		return
	}
	if o.Path != "" {
		query.Set("path", o.Path)
	}
	if o.Sort != "" {
		query.Set("sort", o.Sort)
	}
	if o.Limit > 0 {
		query.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		query.Set("offset", strconv.Itoa(o.Offset))
	}
	if o.PreviewSize != "" {
		query.Set("preview_size", o.PreviewSize)
	}
	if o.PreviewCrop {
		query.Set("preview_crop", "true")
	}
	if len(o.Fields) > 0 {
		query.Set("fields", strings.Join(o.Fields, ","))
	}
}

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string) (*PublicResource, *ErrorResponse) {
	return c.GetMetadataForPublicResourceWithOptions(ctx, public_key, nil)
}

// GetMetadataForPublicResourceWithOptions retrieves metadata for a published
// resource. For published folders, opts controls which nested path is inspected
// and how its contents are sorted and paginated.
func (c *Client) GetMetadataForPublicResourceWithOptions(ctx context.Context, publicKey string, opts *PublicResourceOptions) (*PublicResource, *ErrorResponse) {
	if len(publicKey) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	query := url.Values{}
	query.Set("public_key", publicKey)
	opts.apply(query)

	return requestJSON[PublicResource](ctx, c, GET, "public/resources?"+query.Encode(), nil)
}

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	return c.GetDownloadURLForPublicResourceAt(ctx, public_key, "")
}

// GetDownloadURLForPublicResourceAt returns a download link for a resource
// inside a published folder. Pass an empty path to download the published
// resource itself.
func (c *Client) GetDownloadURLForPublicResourceAt(ctx context.Context, publicKey, path string) (*Link, *ErrorResponse) {
	if len(publicKey) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	query := url.Values{}
	query.Set("public_key", publicKey)
	if path != "" {
		query.Set("path", path)
	}

	return requestJSON[Link](ctx, c, GET, "public/resources/download?"+query.Encode(), nil)
}

// SavePublicResourceOptions contains optional parameters for saving a published
// resource to the authenticated user's Downloads folder.
type SavePublicResourceOptions struct {
	Path string // Path to a resource inside a published folder
	Name string // Name to save the resource under
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	return c.SavePublicResourceWithOptions(ctx, public_key, nil)
}

// SavePublicResourceWithOptions copies a published resource into the user's
// Downloads folder, optionally selecting a nested path and a target name.
//
// The API answers 202 with a link to an asynchronous operation when the copy
// runs in the background, and 201 with a link to the created resource otherwise.
func (c *Client) SavePublicResourceWithOptions(ctx context.Context, publicKey string, opts *SavePublicResourceOptions) (*Link, *ErrorResponse) {
	if len(publicKey) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	query := url.Values{}
	query.Set("public_key", publicKey)
	if opts != nil {
		if opts.Path != "" {
			query.Set("path", opts.Path)
		}
		if opts.Name != "" {
			query.Set("name", opts.Name)
		}
	}

	return requestJSON[Link](ctx, c, POST, "public/resources/save-to-disk?"+query.Encode(), nil,
		http.StatusOK, http.StatusCreated, http.StatusAccepted)
}
