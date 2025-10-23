package openapi //nolint:testpackage // access internal helpers

import "testing"

func TestRenderSchema(t *testing.T) {
	t.Run("nil spec", func(t *testing.T) {
		out, err := renderSchema(nil)
		if err != nil {
			t.Fatalf("renderSchema returned error: %v", err)
		}

		if out != "" {
			t.Fatalf("expected empty output for nil spec, got %q", out)
		}
	})

	t.Run("renders spec", func(t *testing.T) {
		spec := &TypeSpec{Type: "string"}

		out, err := renderSchema(spec)
		if err != nil {
			t.Fatalf("renderSchema returned error: %v", err)
		}

		if out != "type: string" {
			t.Fatalf("unexpected rendered YAML: %q", out)
		}
	})
}
