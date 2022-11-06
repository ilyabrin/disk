package disk

import (
	"net/http"
)

func haveError(err error) bool {
	return err != nil
}

func InArray[T comparable](el T, a []T) bool {
	for _, b := range a {
		if b == el {
			return true
		}
	}
	return false
}

// handleResponseCode - API defined http codes
func handleResponseCode(code int) *ErrorResponse {
	if !InArray(code, []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusNotAcceptable,
		http.StatusConflict,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusLocked,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
		http.StatusInsufficientStorage,
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

// jsonDecodeError - JSON encode/decode error
func jsonDecodeError(err error) *ErrorResponse {
	return &ErrorResponse{
		Message:     "JSON Decode Error",
		Description: "error occurred while API response decode",
		StatusCode:  -1,
		Error:       err,
	}
}
