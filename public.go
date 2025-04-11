package disk

import (
	"context"
	"encoding/json"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string) (*PublicResource, *ErrorResponse) {
	if public_key == "" {
		return nil, nil
	}

	resource, err := doRequest[*PublicResource](ctx, c, GET, "public/resources?public_key="+public_key, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return resource, nil
}

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	if public_key == "" {
		return nil, nil
	}

	link, err := doRequest[*Link](ctx, c, GET, "public/resources/download?public_key="+public_key, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	if public_key == "" {
		return nil, nil
	}

	link, err := doRequest[*Link](ctx, c, POST, "public/resources/save-to-disk?public_key="+public_key, nil)
	if err != nil {
		var errorResponse *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
			return nil, errorResponse
		}
		return nil, nil
	}

	return link, nil
}
