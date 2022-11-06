package disk

import (
	"context"
	"encoding/json"
)

type DiskService service

func (s *DiskService) Info(ctx context.Context, params *QueryParams) (*Disk, *ErrorResponse) {
	var disk *Disk
	resp, err := s.client.get(ctx, s.client.apiURL, params)
	if err != nil {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&disk)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	return disk, nil
}
