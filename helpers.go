package disk

import "log"

// handleError logs the error without terminating the program.
func handleError(err error) {
	if err != nil {
		log.Println("Error:", err)
	}
}

// inArray checks if an integer exists in a slice.
func inArray(n int, array []int) bool {
	for _, b := range array {
		if b == n {
			return true
		}
	}
	return false
}
