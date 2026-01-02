package openapi //nolint:testpackage // access internal helpers

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

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

func TestIsBinary(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, *stringSpec}}
	oneOfSpec := &TypeSpec{OneOf: []TypeSpec{*stringSpec, *binarySpec}}
	arraySpec := &TypeSpec{Type: "array", Items: binarySpec}
	nestedArraySpec := &TypeSpec{Type: "array", Items: &TypeSpec{Type: "array", Items: binarySpec}}

	if !isBinary(binarySpec) {
		t.Error("expected binarySpec to be binary")
	}

	if isBinary(stringSpec) {
		t.Error("expected stringSpec to NOT be binary")
	}

	if !isBinary(unionSpec) {
		t.Error("expected unionSpec to be binary")
	}

	if !isBinary(oneOfSpec) {
		t.Error("expected oneOfSpec to be binary")
	}

	if !isBinary(arraySpec) {
		t.Error("expected arraySpec to be binary")
	}

	if !isBinary(nestedArraySpec) {
		t.Error("expected nestedArraySpec to be binary")
	}

	if isBinary(nil) {
		t.Error("expected nil to NOT be binary")
	}
}

func TestIsPureBinary(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, {Type: "string"}}}

	if !isPureBinary(binarySpec) {
		t.Error("expected binarySpec to be pure binary")
	}

	if isPureBinary(unionSpec) {
		t.Error("expected unionSpec to NOT be pure binary")
	}

	if isPureBinary(nil) {
		t.Error("expected nil to NOT be pure binary")
	}
}

func TestIsNotBinary(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}

	if isNotBinary(binarySpec) {
		t.Error("expected binarySpec NOT to be isNotBinary")
	}

	if !isNotBinary(stringSpec) {
		t.Error("expected stringSpec to be isNotBinary")
	}
}

func TestSimplifyJSONBasic(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}

	if simplifyJSON(nil) != nil {
		t.Error("expected nil for nil input")
	}

	if simplifyJSON(binarySpec) != nil {
		t.Error("expected nil for pure binary spec in JSON")
	}
}

func TestSimplifyJSON_AnyOf(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, *stringSpec}}
	multiUnionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, *stringSpec, {Type: "integer"}}}

	got := simplifyJSON(unionSpec)
	if got == nil || got.Type != "string" || got.Format != "" {
		t.Errorf("expected union to simplify to stringSpec, got %#v", got)
	}

	gotMulti := simplifyJSON(multiUnionSpec)
	if gotMulti == nil || len(gotMulti.AnyOf) != 2 {
		t.Errorf("expected multiUnion to simplify to 2 parts, got %#v", gotMulti)
	}
}

func TestSimplifyJSON_OneOf(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}

	oneOfSpec := &TypeSpec{OneOf: []TypeSpec{*binarySpec, *stringSpec}}

	gotOneOf := simplifyJSON(oneOfSpec)
	if gotOneOf == nil || gotOneOf.Type != "string" || gotOneOf.Format != "" {
		t.Errorf("expected oneOf to simplify to stringSpec, got %#v", gotOneOf)
	}

	multiUnionOneOfSpec := &TypeSpec{OneOf: []TypeSpec{*binarySpec, *stringSpec, {Type: "integer"}}}

	gotMultiOneOf := simplifyJSON(multiUnionOneOfSpec)
	if gotMultiOneOf == nil || len(gotMultiOneOf.OneOf) != 2 {
		t.Errorf("expected multiUnionOneOf to simplify to 2 parts, got %#v", gotMultiOneOf)
	}
}

func TestSimplifyJSONArray(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, *stringSpec}}
	arraySpec := &TypeSpec{Type: "array", Items: unionSpec}

	gotArray := simplifyJSON(arraySpec)
	if gotArray == nil || gotArray.Items == nil || gotArray.Items.Type != "string" {
		t.Errorf("expected array items to be simplified, got %#v", gotArray)
	}
}

func TestSimplifyMultipart(t *testing.T) {
	binarySpec := &TypeSpec{Type: "string", Format: "binary"}
	stringSpec := &TypeSpec{Type: "string"}
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{*binarySpec, *stringSpec}}

	if simplifyMultipart(nil) != nil {
		t.Error("expected nil for nil input")
	}
	
got := simplifyMultipart(unionSpec)
	if got == nil || got.Format != "binary" {
		t.Errorf("expected union to simplify to binary in multipart, got %#v", got)
	}

	arraySpec := &TypeSpec{Type: "array", Items: unionSpec}

	gotArray := simplifyMultipart(arraySpec)
	if gotArray == nil || gotArray.Items == nil || gotArray.Items.Format != "binary" {
		t.Errorf("expected array items to be simplified in multipart, got %#v", gotArray)
	}

	if simplifyMultipart(stringSpec) != stringSpec {
		t.Error("expected non-binary spec to remain unchanged")
	}
}

func TestRenderSpecializedSchemas(t *testing.T) {
	unionSpec := &TypeSpec{AnyOf: []TypeSpec{{Type: "string", Format: "binary"}, {Type: "string"}}}

	jsonOut, _ := renderJSONSchema(unionSpec)
	if jsonOut != "type: string" {
		t.Errorf("unexpected JSON schema: %q", jsonOut)
	}

	multipartOut, _ := renderMultipartSchema(unionSpec)
	if !strings.Contains(multipartOut, "format: binary") {
		t.Errorf("unexpected multipart schema: %q", multipartOut)
	}
}

func TestRenderTemplate(t *testing.T) {
	data := &TemplateData{
		Title:   "Test API",
		Version: "1.2.3",
		Methods: []Method{
			{
				Name:        "getMe",
				Description: []string{"Description for getMe"},
			},
			{
				Name:        "sendMessage",
				Description: []string{"Description for sendMessage"},
				Params: []MethodParam{
					{
						Name:        "chat_id",
						Description: "Chat ID",
						Required:    true,
						Schema:      &TypeSpec{Type: "integer"},
					},
					{
						Name:        "photo",
						Description: "Photo",
						Required:    true,
						Schema:      &TypeSpec{OneOf: []TypeSpec{{Type: "string", Format: "binary"}, {Type: "string"}}},
					},
				},
				SupportsMultipart: true,
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderTemplate(&buf, data); err != nil {
		t.Fatalf("RenderTemplate returned error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "title: \"Test API\"") {
		t.Errorf("expected title in output")
	}

	// sendMessage application/json should have photo as string
	matched, err := regexp.MatchString(`photo:\s*\n\s+type: string`, out)
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}

	if !matched {
		t.Errorf("expected photo as string in JSON section. Output:\n%s", out)
	}

	// sendMessage multipart/form-data should have photo as binary
	// YAML fields are sorted alphabetically: format before type
	if !strings.Contains(out, "format: binary") || !strings.Contains(out, "type: string") {
		t.Errorf("expected photo as binary in multipart section. Output:\n%s", out)
	}
}