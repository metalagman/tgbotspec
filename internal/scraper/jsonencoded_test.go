package scraper //nolint:testpackage // test covers unexported helper

import "testing"

func TestIsJSONSerializedDescription(t *testing.T) {
	tests := []struct {
		desc string
		want bool
	}{
		{"A JSON-serialized object for an inline keyboard", true},
		{"A JSON-serialized list of special entities", true},
		{"A JSON-serialized Array of options", true},
		{"Optional. A JSON-serialized object for an inline keyboard", false},
		{"Unique identifier for the target chat", false},
	}

	for _, tc := range tests {
		if got := isJSONSerializedDescription(tc.desc); got != tc.want {
			t.Fatalf("isJSONSerializedDescription(%q) = %v, want %v", tc.desc, got, tc.want)
		}
	}
}
