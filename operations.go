package main

import (
	"context"
	"encoding/json"
	"log"
)

func (c *Client) GetOperationStatus(ctx context.Context, operationID string) (*OperationStatus, *ErrorResponse) {

	if len(operationID) < 1 {
		return nil, nil
	}

	var operationStatus *OperationStatus
	var errorResponse *ErrorResponse
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "disk/operations/"+operationID, nil)
	if haveError(err) {
		log.Fatal("Request is not performed")
		return nil, nil
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if err != nil {
			log.Fatal(err)
			return nil, nil // json.Decode error
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&operationStatus); err != nil {
		log.Fatal(err)
		return nil, nil // json.Decode error
	}
	return operationStatus, nil
}
