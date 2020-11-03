package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
)

func (c *Client) DeleteResource(ctx context.Context, path string, permanently bool) error {
	if len(path) < 1 {
		return errors.New("delete error")
	}

	var url string

	// todo: make it better
	if permanently {
		url = "disk/resources?path=" + path + "&permanent=true"
	} else {
		url = "disk/resources?path=" + path + "&permanent=false"
	}

	resp, err := c.doRequest(ctx, DELETE, url, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	checkStatusCode(resp.StatusCode)

	return nil
}

func (c *Client) GetMetadata(path string)    {} // get
func (c *Client) UpdateResource(path string) {} // patch

// CreateDir creates a new dorectory with 'path'(string) name
// todo: can't create nested dirs like newDir/subDir/anotherDir
func (c *Client) CreateDir(ctx context.Context, path string) *Link {
	if len(path) < 1 {
		return nil
	}
	var link *Link

	resp, _ := c.doRequest(ctx, PUT, "disk/resources?path="+path, nil)

	decoded := json.NewDecoder(resp.Body)
	// decoded.DisallowUnknownFields()

	if err := decoded.Decode(&link); err != nil {
		log.Fatal(err)
		return nil
	}
	// todo: if link == nil return ErrorResponse
	return link
}

func (c *Client) CopyResource(path string)                  {} // post
func (c *Client) GetDownloadURL(path string)                {} // get
func (c *Client) GetSortedFiles(path string, sortBy string) {} // get | sortBy = [name = default, uploadDate]
func (c *Client) MoveResource(path string)                  {} // post
func (c *Client) GetPublicResources(path string)            {} // get
func (c *Client) PublishResource(path string)               {} // put
func (c *Client) UnpublishResource(path string)             {} // put
func (c *Client) GetLinkForUpload(path string)              {} // get
func (c *Client) UploadFile(url string)                     {} // post
