package main

import (
	"context"
	"encoding/json"
	"log"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string) (*PublicResource, *ErrorResponse) {
	var resource *PublicResource
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "disk/public/resources?public_key="+public_key, nil)
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

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	var link *Link
	var errorResponse *ErrorResponse
	var err error
	var decoded *json.Decoder

	resp, err := c.doRequest(ctx, GET, "disk/public/resources/download?public_key="+public_key, nil)
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

func (c *Client) SavePublicResource(ctx context.Context, public_key string) {}
