package disk

import (
	"context"
	"encoding/json"
	"net/http"
)

type PublicService service

func (s *PublicService) Meta(ctx context.Context, public_key string, params *QueryParams) (*PublicResource, *ErrorResponse) {
	var resource *PublicResource

	resp, err := s.client.get(ctx, s.client.apiURL+"public/resources?public_key="+public_key, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}

func (s *PublicService) DownloadURL(ctx context.Context, public_key string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.get(ctx, s.client.apiURL+"public/resources/download?public_key="+public_key, params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (s *PublicService) Save(ctx context.Context, public_key string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.post(ctx, s.client.apiURL+"public/resources/save-to-disk?public_key="+public_key, nil, nil, params)
	if haveError(err) || !InArray(resp.StatusCode, []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
	}) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}
