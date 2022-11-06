package disk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type ResourceService service

func (s *ResourceService) Delete(ctx context.Context, path string, permanent bool, params *QueryParams) *ErrorResponse {
	url := "resources?path=" + path + "&permanent=false"
	if permanent {
		url = "resources?path=" + path + "&permanent=true"
	}

	resp, err := s.client.delete(ctx, s.client.apiURL+url, nil, params)
	if haveError(err) {
		return handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}

func (s *ResourceService) Meta(ctx context.Context, path string, params *QueryParams) (*Resource, *ErrorResponse) {
	var resource *Resource

	resp, err := s.client.get(ctx, s.client.apiURL+"resources?path="+path, params)
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

func (s *ResourceService) UpdateMeta(ctx context.Context, path string, custom_properties *Metadata) (*Resource, *ErrorResponse) {
	var resource *Resource
	var body []byte

	body, err := json.Marshal(custom_properties)
	if haveError(err) {
		return nil, jsonDecodeError(err)
	}

	resp, err := s.client.patch(ctx, s.client.apiURL+"resources?path="+path, bytes.NewBuffer(body), nil, nil)
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

// CreateDir creates a new dorectory with 'path'(string) name
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (s *ResourceService) CreateDir(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.put(ctx, s.client.apiURL+"resources?path="+path, nil, nil, params)
	if haveError(err) || resp.StatusCode != http.StatusCreated {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&link)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return link, nil
}

func (s *ResourceService) Copy(ctx context.Context, from, to string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.post(ctx, s.client.apiURL+"resources/copy?from="+from+"&path="+to, nil, nil, params)
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

func (s *ResourceService) DownloadURL(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.get(ctx, s.client.apiURL+"resources/download?path="+path, params)
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

// TODO: rename to ListFiles
func (s *ResourceService) GetSortedFiles(ctx context.Context, params *QueryParams) (*FilesResourceList, *ErrorResponse) {
	var files *FilesResourceList

	resp, err := s.client.get(ctx, s.client.apiURL+"resources/files", params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

// get | sortBy = [name = default, uploadDate]
func (s *ResourceService) ListLastUploaded(ctx context.Context, params *QueryParams) (*LastUploadedResourceList, *ErrorResponse) {
	var files *LastUploadedResourceList

	resp, err := s.client.get(ctx, s.client.apiURL+"resources/last-uploaded", params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return files, nil
}

func (s *ResourceService) Move(ctx context.Context, from, to string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.post(ctx, s.client.apiURL+"resources/move?from="+from+"&path="+to, nil, nil, params)
	if haveError(err) || !InArray(resp.StatusCode, []int{
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

func (s *ResourceService) ListPublic(ctx context.Context, params *QueryParams) (*PublicResourcesList, *ErrorResponse) {
	var list *PublicResourcesList

	resp, err := s.client.get(ctx, s.client.apiURL+"resources/public", params)
	if haveError(err) || resp.StatusCode != http.StatusOK {
		return nil, handleResponseCode(resp.StatusCode)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return nil, jsonDecodeError(err)
	}

	return list, nil
}

func (s *ResourceService) Publish(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.put(ctx, s.client.apiURL+"resources/publish?path="+path, nil, nil, params)
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

func (s *ResourceService) Unpublish(ctx context.Context, path string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	resp, err := s.client.put(ctx, s.client.apiURL+"resources/unpublish?path="+path, nil, nil, params)
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

func (s *ResourceService) GetUploadLink(ctx context.Context, path string) (*Link, *ErrorResponse) {
	var resource *Link

	resp, err := s.client.get(ctx, s.client.apiURL+"resources/upload?path="+path, nil)
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

func (s *ResourceService) Upload(ctx context.Context, file, url string, params *QueryParams) *ErrorResponse {
	var errorResponse *ErrorResponse

	f, err := os.Open(file)
	if err != nil {
		return jsonDecodeError(err)
	}
	body := bufio.NewReader(f)
	defer f.Close()

	headers := &HTTPHeaders{
		"Content-Type": "multipart/form-data",
	}
	resp, err := s.client.put(ctx, url, body, headers, nil)
	if haveError(err) {
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		if err != nil {
			return jsonDecodeError(err)
		}
	}
	defer resp.Body.Close()

	return handleResponseCode(resp.StatusCode)
}

func (s *ResourceService) UploadFromURL(ctx context.Context, path, url string, params *QueryParams) (*Link, *ErrorResponse) {
	var link *Link

	ext := filepath.Ext(url) // TODO: fix for files without extension (e.g. www.example.com/filename)
	reqURL := fmt.Sprintf("resources/upload?path=%s%s&url=%s", path, ext, url)
	resp, err := s.client.post(ctx, s.client.apiURL+reqURL, nil, nil, params)
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
