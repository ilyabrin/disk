package main

import (
	"encoding/json"
	"fmt"
)

func prettyPrint(data interface{}) []byte {
	result, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	return result
}

func checkStatusCode(code int) {
	fmt.Printf("\n\n http.StatusCode is: %d \n\n", code)
}

// fmt.Println(humanize.Bytes(uint64(disk.MaxFileSize)))

func jsonErrorResponse(httpStatusCode int) *ErrorResponse {
	return possibleErrorResponses[httpStatusCode]
}
