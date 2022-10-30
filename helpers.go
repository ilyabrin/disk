package disk

import (
	"log"
)

func haveError(err error) bool {
	if err != nil {
		log.Fatal(err)
		return true
	}
	return false
}

func inArray(n int, array []int) bool {
	for _, b := range array {
		if b == n {
			return true
		}
	}
	return false
}
