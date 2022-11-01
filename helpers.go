package disk

import (
	"log"
	"net/http"
)

func haveError(err error) bool {
	if err != nil {
		log.Fatal(err)
		return true
	}
	return false
}

// TODO: use generic-based code (for ints and strings)
func inArray(n int, array []int) bool {
	for _, b := range array {
		if b == n {
			return true
		}
	}
	return false
}

// API defined http codes
func handleResponseCode(code int) *ErrorResponse {
	if !inArray(code, []int{
		200, 201, 202, 301, 302, 400, 401, 404, 406, 409, 412, 413, 423, 429, 500, 503, 507,
	}) {
		return &ErrorResponse{
			Message:    "Unknown error",
			StatusCode: -1,
		}
	}
	return &ErrorResponse{
		Message:    http.StatusText(code),
		StatusCode: code,
	}
}

// JSON encode/decode error
func returnDecodeError(err error) *ErrorResponse {
	return &ErrorResponse{
		Message:     "JSON Decode Error",
		Description: "error occurred while API response decode",
		StatusCode:  -1,
		Error:       err,
	}
}
