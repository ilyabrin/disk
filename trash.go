package disk

import (
	"context"
	"encoding/json"
	"net/http"
)

type TrashService service

func (s *TrashService) Delete(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	resp, err := s.client.delete(ctx, s.client.apiURL+"trash/resources?path="+path, nil, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	var link *Link

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&link)
		if err != nil {
			return nil, jsonDecodeError(err)
		}
	}

	return nil, nil
}

// RestoreFromTrash -
func (s *TrashService) Restore(ctx context.Context, path string, params *QueryParams) (*Link, *Operation, *ErrorResponse) {
	var link *Link

	resp, err := s.client.put(ctx, s.client.apiURL+"trash/resources/restore?path="+path, nil, nil, params)
	if haveError(err) {
		return nil, nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if haveError(err) {
		return nil, nil, jsonDecodeError(err)
	}

	return link, nil, nil
}

// ListTrashResources -
func (s *TrashService) List(ctx context.Context, path string, params *QueryParams) (*TrashResource, *ErrorResponse) {
	var resource *TrashResource

	resp, err := s.client.get(ctx, s.client.apiURL+"trash/resources?path="+path, params)
	if haveError(err) {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	return resource, nil
}
