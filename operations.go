package disk

import (
	"context"
	"encoding/json"
	"fmt"
)

// OperationStatus retrieves the status of an operation by its ID.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and timeouts.
//   - operationID: The unique identifier of the operation to retrieve the status for.
//
// Returns:
//   - *Operation: The operation details if the request is successful.
//   - *ErrorResponse: The error response if the request fails and the error can be unmarshaled.
//   - error: An error if the operation ID is empty or if the request fails and the error cannot be unmarshaled.
//
// Notes:
//   - If the operationID is empty, the function returns an error.
//   - If the request fails and the error response can be unmarshaled, it is returned as *ErrorResponse.
//   - If the request fails and the error response cannot be unmarshaled, a wrapped error is returned.
func (c *Client) OperationStatus(ctx context.Context, operationID string) (*Operation, *ErrorResponse, error) {
	if operationID == "" {
		return nil, nil, fmt.Errorf("operation ID cannot be empty")
	}

	operation, err := doRequest[*Operation](ctx, c, GET, fmt.Sprintf("operations/operation_id=%s", operationID), nil)
	if err != nil {
		var errorResp *ErrorResponse
		if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResp); jsonErr == nil {
			return nil, errorResp, nil
		}
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}

	return operation, nil, nil
}
