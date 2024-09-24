package disk

import "log"

// handleError is a helper function to handle errors
// and exit the program if an error occurs
func handleError(err error) {
	if err != nil {
		log.Fatal("Error:", err)
	}
}

func inArray(n int, array []int) bool {
	if len(array) == 0 {
		return false
	}

	set := make(map[int]struct{}, len(array))
	for _, b := range array {
		set[b] = struct{}{}
	}

	_, exists := set[n]
	return exists
}
