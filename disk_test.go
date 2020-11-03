package main

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TEST_DATA_DIR = "testdata/responses/"

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return cli, s.Close
}

func loadTestResponse(actionName string) []byte {
	response, _ := ioutil.ReadFile(TEST_DATA_DIR + "disk.json")
	return response
}

func TestClientGetDiskInfo(t *testing.T) {

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))
		w.Write(loadTestResponse("disk"))
	})

	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	client := New("token")
	client.HTTPClient = httpClient

	disk, err := client.DiskInfo(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, true, disk.IsPaid)
}
