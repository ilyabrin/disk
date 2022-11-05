package disk

import (
	"context"
	"encoding/json"
)

type OperationService service

func (s *OperationService) Status(ctx context.Context, operationID string, params *QueryParams) (*Operation, *ErrorResponse) {
	resp, err := s.client.get(ctx, s.client.apiURL+"operations/"+operationID, params)
	if err != nil { // || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	operation := new(Operation)
	err = json.NewDecoder(resp.Body).Decode(&operation)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return operation, nil
}
