package disk

import (
	"testing"
)

func TestInArray(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		array    []int
		expected bool
	}{
		{
			name:     "Empty array",
			n:        5,
			array:    []int{},
			expected: false,
		},
		{
			name:     "Single element array, element present",
			n:        5,
			array:    []int{5},
			expected: true,
		},
		{
			name:     "Single element array, element not present",
			n:        5,
			array:    []int{3},
			expected: false,
		},
		{
			name:     "Multiple elements, element present",
			n:        5,
			array:    []int{1, 2, 3, 4, 5, 6, 7},
			expected: true,
		},
		{
			name:     "Multiple elements, element not present",
			n:        10,
			array:    []int{1, 2, 3, 4, 5, 6, 7},
			expected: false,
		},
		{
			name:     "Duplicate elements, element present",
			n:        5,
			array:    []int{1, 5, 2, 5, 3, 5, 4},
			expected: true,
		},
		{
			name: "Large array, element present",
			n:    999,
			array: func() []int {
				arr := make([]int, 1000)
				for i := range arr {
					arr[i] = i
				}
				return arr
			}(),
			expected: true,
		},
		{
			name: "Large array, element not present",
			n:    1000,
			array: func() []int {
				arr := make([]int, 1000)
				for i := range arr {
					arr[i] = i
				}
				return arr
			}(),
			expected: false,
		},
		{
			name:     "Negative numbers, element present",
			n:        -5,
			array:    []int{-7, -6, -5, -4, -3},
			expected: true,
		},
		{
			name:     "Negative numbers, element not present",
			n:        -8,
			array:    []int{-7, -6, -5, -4, -3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inArray(tt.n, tt.array)
			if result != tt.expected {
				t.Errorf("inArray(%d, %v) = %v; want %v", tt.n, tt.array, result, tt.expected)
			}
		})
	}
}

