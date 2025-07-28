package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// TODO: add tests and use generics instead of interface{}
func (c *Client) OperationStatus(ctx context.Context, operationID string) (interface{}, *http.Response, error) {
	query := url.Values{}
	query.Set("operation_id", operationID)
	resp, err := c.doRequest(ctx, GET, "operations?"+query.Encode(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, resp, fmt.Errorf("failed to decode error response: %w", err)
		}
		return &errorResp, resp, nil
	}

	var operation Operation
	if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
		return nil, resp, fmt.Errorf("failed to decode operation: %w", err)
	}

	return &operation, resp, nil
}
