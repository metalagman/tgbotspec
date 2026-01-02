package parser //nolint:testpackage // tests verify internal helpers

import (
	"errors"
	"reflect"
	"testing"
)

//nolint:cyclop // assertions for fields increase complexity
func TestParseType(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h3>Available types</h3>
		<h4><a class="anchor" name="mytype"></a>MyType</h4>
		<p>The primary description.</p>
		<ul><li>Bullet item</li></ul>
		<ol><li>Ordered item</li></ol>
		<table>
		  <tbody>
		    <tr><td> </td><td>String</td><td>Ignored</td></tr>
		    <tr><td>chat_id</td><td>String</td><td>Optional. Target chat</td></tr>
		    <tr><td>count</td><td>Integer</td><td>Number of items</td></tr>
		    <tr><td>target_chat_id</td><td>Integer or String</td><td>Optional. Target chat</td></tr>
		  </tbody>
		</table>
		<blockquote><p>Note text.</p></blockquote>
		<h4>Other</h4>
		</body></html>
	`)

	typeDef, err := ParseType(doc, "mytype")
	if err != nil {
		t.Fatalf("ParseType returned error: %v", err)
	}

	if typeDef.Tag != "Available types" {
		t.Fatalf("expected tag, got %q", typeDef.Tag)
	}

	if got := len(typeDef.Description); got != 3 {
		t.Fatalf("expected three description entries, got %d", got)
	}

	if got := len(typeDef.Fields); got != 3 {
		t.Fatalf("expected three fields, got %d", got)
	}

	fields := typeDef.Fields
	if fields[0].Name != "chat_id" || fields[0].TypeRef.RawType != "Integer" || fields[0].Required {
		t.Fatalf("chat_id field not normalized: %#v", fields[0])
	}

	if !fields[1].Required {
		t.Fatalf("expected second field to be required")
	}

	if fields[2].Name != "target_chat_id" || fields[2].TypeRef.RawType != "Integer" || fields[2].Required {
		t.Fatalf("target_chat_id field not normalized: %#v", fields[2])
	}

	if !reflect.DeepEqual(typeDef.Notes, []string{"Note text."}) {
		t.Fatalf("unexpected notes: %#v", typeDef.Notes)
	}
}

func TestParseTypeMissingAnchor(t *testing.T) {
	doc := mustDoc(t, `<html><body><h4><a class="anchor" name="other"></a>Other</h4></body></html>`)

	_, err := ParseType(doc, "missing")
	if !errors.Is(err, ErrElementNotFound) {
		t.Fatalf("expected ErrElementNotFound, got %v", err)
	}
}

func TestParseTypeStopsAtNextHeader(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h4><a class="anchor" name="simple"></a>Simple</h4>
		<p>First description.</p>
		<h4><a class="anchor" name="second"></a>Second</h4>
		<table><tbody><tr><td>name</td><td>String</td><td>Optional info</td></tr></tbody></table>
		</body></html>
	`)

	typeDef, err := ParseType(doc, "simple")
	if err != nil {
		t.Fatalf("ParseType returned error: %v", err)
	}

	if len(typeDef.Description) != 1 {
		t.Fatalf("expected description to stop before next header, got %#v", typeDef.Description)
	}
}

//nolint:cyclop,funlen // exhaustive assertions for union parsing
func TestTypeRefUnionAndSpec(t *testing.T) {
	if parts := NewTypeRef("Sticker or Photo").UnionParts(); !reflect.DeepEqual(parts, []string{"Sticker", "Photo"}) {
		t.Fatalf("unexpected union parts: %#v", parts)
	}

	if parts := NewTypeRef("Sticker and Photo").UnionParts(); !reflect.DeepEqual(parts, []string{"Sticker", "Photo"}) {
		t.Fatalf("unexpected union parts for and: %#v", parts)
	}

	if parts := NewTypeRef("One, Two and Three").UnionParts(); !reflect.DeepEqual(parts, []string{"One", "Two", "Three"}) {
		t.Fatalf("unexpected union parts for comma list: %#v", parts)
	}

	if parts := NewTypeRef("One, , Three").UnionParts(); !reflect.DeepEqual(parts, []string{"One", "Three"}) {
		t.Fatalf("expected blank entries skipped, got %#v", parts)
	}

	if NewTypeRef("Sticker").UnionParts() != nil {
		t.Fatalf("expected nil union for single value")
	}

	if NewTypeRef(" ").UnionParts() != nil {
		t.Fatalf("expected nil union for blank input")
	}

	if NewTypeRef("One or ").UnionParts() != nil {
		t.Fatalf("expected nil union when second part missing")
	}

	if NewTypeRef("One, Two").UnionParts() == nil {
		t.Fatalf("expected union for comma-separated pair")
	}

	spec := NewTypeRef("Array of Message").ToTypeSpec()
	if spec.Type != "array" || spec.Items == nil || spec.Items.Ref == nil || spec.Items.Ref.Name != "Message" {
		t.Fatalf("array spec unexpected: %#v", spec)
	}

	spec = NewTypeRef("Array of Array of Sticker").ToTypeSpec()
	if spec.Type != "array" || spec.Items == nil || spec.Items.Type != "array" {
		t.Fatalf("nested array spec unexpected: %#v", spec)
	}

	spec = NewTypeRef("array of array of Sticker").ToTypeSpec()
	if spec.Type != "array" {
		t.Fatalf("expected array type for lowercase phrasing, got %#v", spec)
	}

	spec = NewTypeRef("Integer").ToTypeSpec()
	if spec.Type != "integer" {
		t.Fatalf("primitive spec unexpected: %#v", spec)
	}

	spec = NewTypeRef("int64").ToTypeSpec()
	if spec.Format != "int64" {
		t.Fatalf("expected int64 format, got %#v", spec)
	}

	spec = NewTypeRef("string").ToTypeSpec()
	if spec.Type != "string" {
		t.Fatalf("expected string type, got %#v", spec)
	}

	spec = NewTypeRef("True").ToTypeSpec()
	if spec.Default != true {
		t.Fatalf("expected True default, got %#v", spec)
	}

	spec = NewTypeRef("Sticker or Photo").ToTypeSpec()
	if len(spec.OneOf) != 2 {
		t.Fatalf("expected oneOf for union, got %#v", spec)
	}

	spec = NewTypeRef("CustomType").ToTypeSpec()
	if spec.Ref == nil || spec.Ref.Name != "CustomType" {
		t.Fatalf("expected ref for custom type, got %#v", spec)
	}

	spec = (*TypeRef)(nil).ToTypeSpec()
	if spec.Type != "" || spec.Ref != nil {
		t.Fatalf("expected empty spec for nil TypeRef")
	}

	// Extra cases for parseBasicType
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{"bool", "boolean"},
		{"float", "number"},
		{"number", "number"},
		{"inputfile", "string"},
	} {
		spec = NewTypeRef(tc.raw).ToTypeSpec()
		if spec.Type != tc.want {
			t.Errorf("parseBasicType(%q) = %q, want %q", tc.raw, spec.Type, tc.want)
		}
	}
}

func TestTypeRefContainsType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		target string
		want   bool
	}{
		{name: "nil receiver", raw: "", target: "InputFile", want: false},
		{name: "empty target", raw: "InputFile", target: "  ", want: false},
		{name: "empty raw", raw: " ", target: "InputFile", want: false},
		{name: "direct match", raw: "InputFile", target: "InputFile", want: true},
		{name: "case insensitive", raw: "inputfile", target: "InputFile", want: true},
		{name: "array nesting", raw: "Array of Array of InputFile", target: "InputFile", want: true},
		{name: "union contains", raw: "Message or InputFile", target: "InputFile", want: true},
		{name: "union no match", raw: "Message or Sticker", target: "InputFile", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr *TypeRef
			if tc.raw != "" {
				tr = NewTypeRef(tc.raw)
			}

			if got := tr.ContainsType(tc.target); got != tc.want {
				t.Fatalf("ContainsType(%q) on %q = %v, want %v", tc.target, tc.raw, got, tc.want)
			}
		})
	}
}

func TestTypeRefContainsTypeWithPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		prefix string
		want   bool
	}{
		{name: "empty prefix", raw: "InputMediaPhoto", prefix: "", want: false},
		{name: "direct match", raw: "InputMediaPhoto", prefix: "InputMedia", want: true},
		{name: "array nesting", raw: "Array of InputMediaAnimation", prefix: "InputMedia", want: true},
		{name: "union match", raw: "String or InputStickerAnimated", prefix: "InputSticker", want: true},
		{name: "union no match", raw: "String or Message", prefix: "InputMedia", want: false},
		{name: "case insensitive", raw: "inputmediaphoto", prefix: "inputmedia", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr *TypeRef
			if tc.raw != "" {
				tr = NewTypeRef(tc.raw)
			}

			if got := tr.ContainsTypeWithPrefix(tc.prefix); got != tc.want {
				t.Fatalf("ContainsTypeWithPrefix(%q) on %q = %v, want %v", tc.prefix, tc.raw, got, tc.want)
			}
		})
	}
}

func TestReturnTypeToSpec(t *testing.T) { //nolint:cyclop // sequential assertions cover many cases
	// Array of Update -> array items $ref Update
	tr := NewTypeRef("Array of Update")

	spec := tr.ToTypeSpec()
	if spec.Type != "array" {
		t.Fatalf("expected array type, got %q", spec.Type)
	}

	if spec.Items == nil || spec.Items.Ref == nil || spec.Items.Ref.Name != "Update" {
		t.Fatalf("expected items ref to Update, got %#v", spec.Items)
	}

	// True -> boolean default true
	tr = NewTypeRef("True")

	spec = tr.ToTypeSpec()
	if spec.Type != "boolean" || spec.Default != true {
		t.Fatalf("expected boolean with default true, got type=%q default=%v", spec.Type, spec.Default)
	}

	// WebhookInfo -> $ref
	tr = NewTypeRef("WebhookInfo")

	spec = tr.ToTypeSpec()
	if spec.Ref == nil || spec.Ref.Name != "WebhookInfo" {
		t.Fatalf("expected $ref WebhookInfo, got %#v", spec)
	}

	// StarAmount -> $ref
	tr = NewTypeRef("StarAmount")

	spec = tr.ToTypeSpec()
	if spec.Ref == nil || spec.Ref.Name != "StarAmount" {
		t.Fatalf("expected $ref StarAmount, got %#v", spec)
	}

	tr = NewTypeRef("Boolean")

	spec = tr.ToTypeSpec()
	if spec.Type != "boolean" {
		t.Fatalf("expected boolean type, got %#v", spec)
	}

	tr = NewTypeRef("Float number")

	spec = tr.ToTypeSpec()
	if spec.Type != "number" {
		t.Fatalf("expected number type, got %#v", spec)
	}

	tr = NewTypeRef("Int")

	spec = tr.ToTypeSpec()
	if spec.Type != "integer" {
		t.Fatalf("expected integer type, got %#v", spec)
	}
}
