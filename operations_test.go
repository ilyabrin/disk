package disk

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperationStatus(t *testing.T) {
	t.Run("Successful operation status", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.Header.Get("Authorization"))
				assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

				w.Write([]byte(
					`{
						"status": "success"
					}`))
			}))

		operation, errResp, err := client.OperationStatus(context.Background(), "12345")

		assert.NoError(t, err)
		assert.Nil(t, errResp)
		assert.NotNil(t, operation)
		assert.Equal(t, "success", operation.Status)
	})

	t.Run("Empty operation ID", func(t *testing.T) {
		client := mockedHttpClient(nil)

		operation, errResp, err := client.OperationStatus(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, operation)
		assert.Nil(t, errResp)
		assert.Equal(t, "operation ID cannot be empty", err.Error())
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		client := mockedHttpClient(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.Header.Get("Authorization"))
				assert.Equal(t, "OAuth token", r.Header.Get("Authorization"))

				w.Write([]byte(`invalid-json`))
			}))

		operation, errResp, err := client.OperationStatus(context.Background(), "12345")

		assert.Error(t, err)
		assert.Nil(t, operation)
		assert.Nil(t, errResp)
		assert.Contains(t, err.Error(), "failed to make request")
	})
}
