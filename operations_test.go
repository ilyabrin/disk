package disk

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockRoundTripper is a mock HTTP transport for testing
type mockRoundTripper struct {
	response    *http.Response
	shouldError bool
	errorMsg    string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return m.response, nil
}

// stringReadCloser creates an io.ReadCloser from a string
func stringReadCloser(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func TestOperationStatus(t *testing.T) {
	t.Run("OperationStatus with valid operation ID", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client to return a valid operation response
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				response: &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Header:     make(http.Header),
					Body:       stringReadCloser(`{"status": "completed", "operation_id": "test-op-123"}`),
				},
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "test-op-123")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if resp == nil {
			t.Error("Expected response to be returned")
		}

		if result == nil {
			t.Error("Expected result to be returned")
		}

		// Check if result is an Operation
		if operation, ok := result.(*Operation); ok {
			if operation.Status != "completed" {
				t.Errorf("Expected status 'completed', got '%s'", operation.Status)
			}
		} else {
			t.Error("Expected result to be an Operation")
		}
	})

	t.Run("OperationStatus with error response", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client to return an error response
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Header:     make(http.Header),
					Body:       stringReadCloser(`{"error": "OperationNotFoundError", "description": "Operation not found"}`),
				},
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "non-existent-op")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if resp == nil {
			t.Error("Expected response to be returned")
		}

		if resp.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", resp.StatusCode)
		}

		// Check if result is an ErrorResponse
		if errorResp, ok := result.(*ErrorResponse); ok {
			if errorResp.Error != "OperationNotFoundError" {
				t.Errorf("Expected error 'OperationNotFoundError', got '%s'", errorResp.Error)
			}
		} else {
			t.Error("Expected result to be an ErrorResponse")
		}
	})

	t.Run("OperationStatus with empty operation ID", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				response: &http.Response{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Header:     make(http.Header),
					Body:       stringReadCloser(`{"error": "BadRequestError", "description": "Operation ID is required"}`),
				},
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if resp == nil {
			t.Error("Expected response to be returned")
		}

		if resp.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", resp.StatusCode)
		}

		// Check if result is an ErrorResponse
		if errorResp, ok := result.(*ErrorResponse); ok {
			if !strings.Contains(errorResp.Error, "BadRequestError") {
				t.Errorf("Expected error to contain 'BadRequestError', got '%s'", errorResp.Error)
			}
		} else {
			t.Error("Expected result to be an ErrorResponse")
		}
	})

	t.Run("OperationStatus with network error", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client to return a network error
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				shouldError: true,
				errorMsg:    "network error",
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "test-op-123")
		if err == nil {
			t.Error("Expected error for network failure")
		}

		if !strings.Contains(err.Error(), "failed to make request") {
			t.Errorf("Expected error to contain 'failed to make request', got: %v", err)
		}

		if result != nil {
			t.Error("Expected nil result on network error")
		}

		if resp != nil {
			t.Error("Expected nil response on network error")
		}
	})

	t.Run("OperationStatus with invalid JSON response", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client to return invalid JSON
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				response: &http.Response{
					StatusCode: 200,
					Status:     "200 OK",
					Header:     make(http.Header),
					Body:       stringReadCloser(`{invalid json}`),
				},
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "test-op-123")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}

		if !strings.Contains(err.Error(), "failed to decode operation") {
			t.Errorf("Expected error to contain 'failed to decode operation', got: %v", err)
		}

		if result != nil {
			t.Error("Expected nil result on decode error")
		}

		if resp == nil {
			t.Error("Expected response to be returned even on decode error")
		}
	})

	t.Run("OperationStatus with invalid error JSON response", func(t *testing.T) {
		client, _ := New("test-token")

		// Mock HTTP client to return invalid JSON for error response
		client.HTTPClient = &http.Client{
			Transport: &mockRoundTripper{
				response: &http.Response{
					StatusCode: 404,
					Status:     "404 Not Found",
					Header:     make(http.Header),
					Body:       stringReadCloser(`{invalid error json}`),
				},
			},
		}

		result, resp, err := client.OperationStatus(context.Background(), "test-op-123")
		if err == nil {
			t.Error("Expected error for invalid error JSON")
		}

		if !strings.Contains(err.Error(), "failed to decode error response") {
			t.Errorf("Expected error to contain 'failed to decode error response', got: %v", err)
		}

		if result != nil {
			t.Error("Expected nil result on decode error")
		}

		if resp == nil {
			t.Error("Expected response to be returned even on decode error")
		}
	})
}