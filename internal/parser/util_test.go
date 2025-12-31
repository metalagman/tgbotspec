package parser //nolint:testpackage // tests verify internal helpers

import "testing"

func TestIsOptionalDescription(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"Optional. Extra", true},
		{" optional parameter", true},
		{"Required", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isOptionalDescription(tc.input); got != tc.want {
			t.Fatalf("isOptionalDescription(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
