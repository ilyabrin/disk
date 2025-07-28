package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetMetadataForPublicResource(ctx context.Context, public_key string) (*PublicResource, *ErrorResponse) {
	if len(public_key) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	var resource *PublicResource
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, GET, "public/resources?public_key="+public_key, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&resource); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode resource: %v", err)}
	}
	return resource, nil
}

func (c *Client) GetDownloadURLForPublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	if len(public_key) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, GET, "public/resources/download?public_key="+public_key, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}

func (c *Client) SavePublicResource(ctx context.Context, public_key string) (*Link, *ErrorResponse) {
	if len(public_key) < 1 {
		return nil, &ErrorResponse{Error: "public_key cannot be empty"}
	}

	var link *Link
	var errorResponse *ErrorResponse

	resp, err := c.doRequest(ctx, POST, "public/resources/save-to-disk?public_key="+public_key, nil)
	if err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	// Если сохранение происходит асинхронно,
	// то вернёт ответ с кодом 202 и ссылкой на асинхронную операцию.
	// Иначе вернёт ответ с кодом 201 и ссылкой на созданный ресурс.
	if !inArray(resp.StatusCode, []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
	}) {
		decoded := json.NewDecoder(resp.Body)
		if err := decoded.Decode(&errorResponse); err != nil {
			return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode error response: %v", err)}
		}
		return nil, errorResponse
	}

	decoded := json.NewDecoder(resp.Body)
	if err := decoded.Decode(&link); err != nil {
		return nil, &ErrorResponse{Error: fmt.Sprintf("failed to decode link: %v", err)}
	}

	return link, nil
}
