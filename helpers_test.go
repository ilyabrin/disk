package disk

import "testing"

func Test_inArray(t *testing.T) {
	tests := []struct {
		name   string
		got    any
		expect any
	}{
		{name: "when inArray TRUE", got: inArray(5, []int{1, 2, 4, 6, 4, 5}), expect: true},
		{name: "when inArray FALSE", got: inArray(7, []int{1, 2, 4, 6, 4, 5}), expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, tc.got)
			}
		})
	}
}
