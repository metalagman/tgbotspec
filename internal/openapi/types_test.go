package openapi_test

import (
	"testing"

	"tgbotspec/internal/openapi"
)

func TestTypeRefMarshalYAML(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		var tr openapi.TypeRef

		res, err := tr.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML returned error: %v", err)
		}

		if res != nil {
			t.Fatalf("expected nil result for empty name, got %#v", res)
		}
	})

	t.Run("with name", func(t *testing.T) {
		tr := openapi.TypeRef{Name: "User"}

		res, err := tr.MarshalYAML()
		if err != nil {
			t.Fatalf("MarshalYAML returned error: %v", err)
		}

		str, ok := res.(string)
		if !ok {
			t.Fatalf("expected string result, got %T", res)
		}

		expected := "#/components/schemas/User"
		if str != expected {
			t.Fatalf("expected %q, got %q", expected, str)
		}
	})
}
