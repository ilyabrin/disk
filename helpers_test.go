package disk_test

import (
	"testing"

	"github.com/ilyabrin/disk"
)

func Test_InArray(t *testing.T) {
	tests := []struct {
		name   string
		got    any
		expect any
	}{
		{name: "when InArray TRUE", got: disk.InArray(5, []int{1, 2, 4, 6, 4, 5}), expect: true},
		{name: "when InArray FALSE", got: disk.InArray(7, []int{1, 2, 4, 6, 4, 5}), expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, tc.got)
			}
		})
	}
}
