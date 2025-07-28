package disk

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
