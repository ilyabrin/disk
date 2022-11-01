package disk

import (
	"context"
	"encoding/json"
	"log"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string, params *queryParams) (*PublicResource, *ErrorResponse) {
	var resource *PublicResource
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"public/resources?public_key="+public_key, nil, nil, params)
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

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string, params *queryParams) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "GET", c.api_url+"public/resources/download?public_key="+public_key, nil, nil, params)
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
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string, params *queryParams) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, "POST", c.api_url+"public/resources/save-to-disk?public_key="+public_key, nil, nil, params)
	if haveError(err) {
		log.Fatal("Request failed")
	}

	// Если сохранение происходит асинхронно,
	// то вернёт ответ с кодом 202 и ссылкой на асинхронную операцию.
	// Иначе вернёт ответ с кодом 201 и ссылкой на созданный ресурс.
	if !inArray(resp.StatusCode, []int{200, 201, 202}) {
		decoded = json.NewDecoder(resp.Body)
		err := decoded.Decode(&errorResponse)
		if haveError(err) {
			log.Fatal(err)
		}
		return nil, errorResponse
	}

	decoded = json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
	}

	return link, nil
}
