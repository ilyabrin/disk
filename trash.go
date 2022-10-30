package disk

import (
	"context"
	"encoding/json"
	"log"
)

func (c *Client) DeleteFromTrash(ctx context.Context, path string, params *optional_params) (*Link, *ErrorResponse) {
	var resource *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, DELETE, c.api_url+"trash/resources?path="+path, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode == 204 {
		return nil, nil
	}

	if !inArray(resp.StatusCode, []int{200, 202}) {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	return resource, nil
}

// TODO: refactor ASAP
func (c *Client) RestoreFromTrash(ctx context.Context, path string, params *optional_params) (*Link, *Operation, *ErrorResponse) {
	var link *Link
	// var operation *Operation
	// var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, PUT, c.api_url+"trash/resources/restore?path="+path, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	// if resp.StatusCode != 202 {
	// 	decoded = json.NewDecoder(resp.Body)
	// 	if err := decoded.Decode(&operation); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	return nil, operation, nil
	// }

	// if !inArray(resp.StatusCode, []int{200, 201}) {
	// 	decoded = json.NewDecoder(resp.Body)
	// 	err := decoded.Decode(&errorResponse)
	// 	if haveError(err) {
	// 		log.Fatal(err)
	// 	}
	// 	return nil, nil, errorResponse
	// }

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil, nil
}

func (c *Client) ListTrashResources(ctx context.Context, path string, params *optional_params) (*TrashResource, *ErrorResponse) {
	var resource *TrashResource
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, c.api_url+"trash/resources?path="+path, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	if resp.StatusCode != 200 {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		log.Fatal(err)
	}

	return resource, nil
}
