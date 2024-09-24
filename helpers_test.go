package disk

import (
	"bytes"
	"log"
	"testing"
)

// Mock for os.Exit
var osExitCalled = false
var osExitCode = 0
var osExit = func(code int) {
	osExitCalled = true
	osExitCode = code
	panic("os.Exit called")
}

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

func TestHandleError(t *testing.T) {

	// Save the original log output and flags
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	defer func() {
		// Restore the original log output and flags after the test
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
	}()

	// Create a buffer to capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Remove timestamp from log output for easier testing
	log.SetFlags(0)

	// Override os.Exit to prevent the test from terminating
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called")
	}

	tests := []struct {
		name          string
		err           error
		expectedLog   string
		expectedPanic bool
	}{
		{
			name:          "Nil error",
			err:           nil,
			expectedLog:   "",
			expectedPanic: false,
		},
		// todo
		// {
		// 	name:          "Non-nil error",
		// 	err:           errors.New("test error"),
		// 	expectedLog:   "Error: test error\n",
		// 	expectedPanic: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the buffer before each test
			buf.Reset()
			exitCode = 0

			// Use a function to capture panics
			func() {
				defer func() {
					r := recover()
					if (r != nil) != tt.expectedPanic {
						t.Errorf("handleError() panic = %v, expectedPanic %v", r, tt.expectedPanic)
					}
					if r != nil && exitCode != 1 {
						t.Errorf("Expected exit code 1, got %d", exitCode)
					}
				}()
				handleError(tt.err)
			}()

			// Check the log output
			if got := buf.String(); got != tt.expectedLog {
				t.Errorf("handleError() log = %q, want %q", got, tt.expectedLog)
			}
		})
	}
}
