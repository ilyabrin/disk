package disk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// TODO: add tests and use generics instead of interface{}
func (c *Client) OperationStatus(ctx context.Context, operationID string) (any, *http.Response, error) {
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

// OperationIDFromHref extracts the operation identifier from a link returned by
// an asynchronous API call, e.g.
// "https://cloud-api.yandex.net/v1/disk/operations/123abc" -> "123abc".
// Hrefs that are already bare identifiers are returned unchanged.
func OperationIDFromHref(href string) string {
	if href == "" {
		return ""
	}
	if parsed, err := url.Parse(href); err == nil && parsed.Path != "" {
		href = parsed.Path
	}
	return path.Base(strings.TrimSuffix(href, "/"))
}

// GetOperationStatus is a typed wrapper around OperationStatus. It accepts
// either an operation ID or the full href from an asynchronous response.
func (c *Client) GetOperationStatus(ctx context.Context, operationIDOrHref string) (*Operation, error) {
	id := OperationIDFromHref(operationIDOrHref)
	if id == "" {
		return nil, errors.New("operation id cannot be empty")
	}

	result, _, err := c.OperationStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	switch value := result.(type) {
	case *Operation:
		return value, nil
	case *ErrorResponse:
		return nil, fmt.Errorf("operation status failed: %s: %s", value.Error, value.Description)
	default:
		return nil, fmt.Errorf("unexpected operation status response %T", result)
	}
}

// OperationInProgress is the status the API reports while an asynchronous
// operation is still running.
const OperationInProgress = "in-progress"
