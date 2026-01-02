package openapi_test

import (
	"testing"

	"github.com/metalagman/tgbotspec/internal/openapi"
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

func TestTypeSpecWithDescription(t *testing.T) {
	t.Run("nil spec", func(t *testing.T) {
		var s *openapi.TypeSpec
		if s.WithDescription("desc") != nil {
			t.Fatal("expected nil for nil spec")
		}
	})

	t.Run("simple spec", func(t *testing.T) {
		s := &openapi.TypeSpec{Type: "string"}
		res := s.WithDescription("my description")

		if res.Description != "my description" {
			t.Errorf("expected description to be set, got %q", res.Description)
		}

		if res.Type != "string" {
			t.Errorf("expected type to be preserved, got %q", res.Type)
		}
	})

	t.Run("spec with ref", func(t *testing.T) {
		s := &openapi.TypeSpec{Ref: &openapi.TypeRef{Name: "User"}}
		res := s.WithDescription("user description")

		if res.Description != "user description" {
			t.Errorf("expected description to be set, got %q", res.Description)
		}

		if res.Ref != nil {
			t.Fatal("expected Ref to be nil in the wrapper spec")
		}

		if len(res.AllOf) != 1 || res.AllOf[0].Ref == nil || res.AllOf[0].Ref.Name != "User" {
			t.Errorf("expected AllOf to wrap the reference, got %#v", res.AllOf)
		}
	})
}
