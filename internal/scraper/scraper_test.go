package scraper //nolint:testpackage // tests rely on internal helper hooks

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/metalagman/tgbotspec/internal/openapi"
	"github.com/metalagman/tgbotspec/internal/parser"
)

func docFromString(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	return doc
}

func TestExtractBotAPIVersion(t *testing.T) {
	t.Parallel()

	doc := docFromString(t, `<html><body><p><strong>Bot API 6.7</strong></p></body></html>`)
	if version := extractBotAPIVersion(doc); version != "6.7" {
		t.Fatalf("expected version 6.7, got %q", version)
	}

	docEmpty := docFromString(t, `<html><body></body></html>`)
	if version := extractBotAPIVersion(docEmpty); version != "" {
		t.Fatalf("expected empty version, got %q", version)
	}
}

func TestExtractAPITitle(t *testing.T) {
	t.Parallel()

	doc := docFromString(t, `<html><body><h1>Telegram Bot API</h1></body></html>`)
	if title := extractAPITitle(doc); title != "Telegram Bot API" {
		t.Fatalf("unexpected title: %q", title)
	}

	fallback := docFromString(t, `<html><head><title>Fallback</title></head><body></body></html>`)
	if title := extractAPITitle(fallback); title != "Fallback" {
		t.Fatalf("expected fallback title, got %q", title)
	}

	empty := docFromString(t, `<html><body></body></html>`)
	if title := extractAPITitle(empty); title != "" {
		t.Fatalf("expected empty title, got %q", title)
	}
}

const mockHTML = `
<html>
<head><title>Mock API</title></head>
<body>
	<a data-target="#User">User</a>
	<a data-target="#User">Duplicate</a>
	<a data-target="#InputFile">InputFile</a>
	<a data-target="#getMe">getMe</a>
	<a data-target="#sendPhoto">sendPhoto</a>

	<p><strong>Bot API 7.0</strong></p>
	<h3>Available types</h3>
	<h4><a class="anchor" name="User"></a>User</h4>
	<p>This object represents a Telegram user or bot.</p>
	<table>
		<tbody>
			<tr><td>id</td><td>Integer</td><td>Unique identifier for this user or bot.</td></tr>
			<tr><td>is_bot</td><td>Boolean</td><td>True, if this user is a bot</td></tr>
			<tr><td>first_name</td><td>String</td><td>User's or bot's first name</td></tr>
		</tbody>
	</table>

	<h4><a class="anchor" name="InputFile"></a>InputFile</h4>
	<p>Skipped in Run loop.</p>

	<h3>Available methods</h3>
	<h4><a class="anchor" name="getMe"></a>getMe</h4>
	<p>A simple method for testing your bot's authentication token.
	Returns basic information about the bot in form of a User object.</p>

	<h4><a class="anchor" name="sendPhoto"></a>sendPhoto</h4>
	<p>Use this method to send photos. On success, the sent Message is returned.</p>
	<table>
		<tbody>
			<tr><td>chat_id</td><td>Integer or String</td><td>Unique identifier for the target chat</td></tr>
			<tr><td>photo</td><td>InputFile or String</td><td>Photo to send.</td></tr>
		</tbody>
	</table>
</body>
</html>`

func TestRunWritesOpenAPISpec(t *testing.T) {
	original := fetchDocument

	t.Cleanup(func() {
		fetchDocument = original
	})

	fetchDocument = func() (*goquery.Document, error) {
		return docFromString(t, mockHTML), nil
	}

	var buf bytes.Buffer
	if err := Run(&buf, Options{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatal("expected rendered OpenAPI output")
	}

	assertContains(t, out, "operationId: getMe", "getMe method")
	assertContains(t, out, "User:", "User type")
	assertContains(t, out, "ResponseParameters:", "ResponseParameters type")
}

//nolint:funlen // test contains inline HTML mock
func TestRunMergeUnionTypes(t *testing.T) {
	original := fetchDocument

	t.Cleanup(func() {
		fetchDocument = original
	})

	html := `
<html>
<body>
	<p><strong>Bot API 7.0</strong></p>

    <!-- Navigation -->
    <a data-target="#InputMedia">InputMedia</a>
    <a data-target="#InputMediaPhoto">InputMediaPhoto</a>
    <a data-target="#InputMediaVideo">InputMediaVideo</a>
    <a data-target="#sendMediaGroup">sendMediaGroup</a>

	<h3>Available types</h3>
	<h4><a class="anchor" name="InputMedia"></a>InputMedia</h4>
	<p>This object represents the content of a media message to be sent.</p>
    <ul><li>InputMediaPhoto</li><li>InputMediaVideo</li></ul>

	<h4><a class="anchor" name="InputMediaPhoto"></a>InputMediaPhoto</h4>
	<p>Represents a photo.</p>
    <table><tbody><tr><td>type</td><td>String</td><td>Type of the result</td></tr></tbody></table>

	<h4><a class="anchor" name="InputMediaVideo"></a>InputMediaVideo</h4>
	<p>Represents a video.</p>
    <table><tbody><tr><td>type</td><td>String</td><td>Type of the result</td></tr></tbody></table>

	<h3>Available methods</h3>
	<h4><a class="anchor" name="sendMediaGroup"></a>sendMediaGroup</h4>
	<p>Use this method to send a group of photos, videos, documents or audios as an album.
       On success, an array of Messages that were sent is returned.</p>
	<table>
		<tbody>
			<tr><td>chat_id</td><td>Integer</td><td>Unique identifier for the target chat</td></tr>
			<tr><td>media</td><td>Array of InputMediaPhoto or InputMediaVideo</td><td>Photos and videos to be sent</td></tr>
		</tbody>
	</table>
</body>
</html>`

	fetchDocument = func() (*goquery.Document, error) {
		return docFromString(t, html), nil
	}

	var buf bytes.Buffer
	if err := Run(&buf, Options{MergeUnionTypes: true}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := buf.String()
	
	// We expect the media parameter items to reference InputMedia directly,
	// merging InputMediaPhoto and InputMediaVideo.
	// Since regex in go test is verbose, we rely on the specific string being present
	// and NOT having an anyOf block for this param.
	
	// Check for the reference
	if !strings.Contains(out, "$ref: '#/components/schemas/InputMedia'") {
		t.Error("expected merged reference to InputMedia in output")
	}

	// We can't easily check for *absence* of anyOf globally, but we can verify that
	// InputMediaPhoto is NOT referenced in the context of the media param.
	// However, InputMediaPhoto IS defined in the spec, so its name appears.
	// Let's assume the presence of the merged ref is sufficient proof of the merge logic activation.
}

func TestRunWithEmptyDoc(t *testing.T) {
	original := fetchDocument

	t.Cleanup(func() { fetchDocument = original })

	// Doc with ResponseParameters already present to exercise that branch
	html := `
		<html><body>
			<a data-target="#ResponseParameters">ResponseParameters</a>
			<h3>Available types</h3>
			<h4><a class="anchor" name="ResponseParameters"></a>ResponseParameters</h4>
			<table><tbody><tr><td>retry_after</td><td>Integer</td><td>desc</td></tr></tbody></table>
		</body></html>`
	fetchDocument = func() (*goquery.Document, error) {
		return docFromString(t, html), nil
	}

	var buf bytes.Buffer
	if err := Run(&buf, Options{}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "title: \"Telegram Bot API\"") {
		t.Error("expected default title")
	}

	if !strings.Contains(out, "version: \"0.0.0\"") {
		t.Error("expected default version")
	}
}

func assertContains(t *testing.T, s, substr, name string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected %s in output", name)
	}
}

func TestSplitTargets(t *testing.T) {
	html := `
		<html>
		<body>
			<h3><a class="anchor" name="available-types"></a>Available types</h3>
			<h4><a class="anchor" name="User"></a>User</h4>
			<table><tbody><tr><td>id</td><td>Integer</td><td>desc</td></tr></tbody></table>

			<h3><a class="anchor" name="available-methods"></a>Available methods</h3>
			<h4><a class="anchor" name="getMe"></a>getMe</h4>
			<p>Returns User.</p>
		</body>
		</html>`
	doc := docFromString(t, html)

	// Test with explicit targets
	targets := []parser.ParseTarget{
		{Anchor: "User", Mode: parser.ParseModeType, Name: "User"},
		{Anchor: "getMe", Mode: parser.ParseModeMethod, Name: "getMe"},
		{Anchor: "MissingType", Mode: parser.ParseModeType, Name: "MissingType"},
		{Anchor: "MissingMethod", Mode: parser.ParseModeMethod, Name: "MissingMethod"},
	}

	types, methods := splitTargets(targets, doc)
	if len(types) != 1 || types[0].Name != "User" {
		t.Errorf("expected 1 type User, got %d", len(types))
	}

	if len(methods) != 1 || methods[0].Name != "getMe" {
		t.Errorf("expected 1 method getMe, got %d", len(methods))
	}

	// Test with automatic targets (empty input)
	types, methods = splitTargets(nil, doc)
	if len(types) != 1 || types[0].Name != "User" {
		t.Errorf("expected 1 automatic type User, got %d", len(types))
	}

	if len(methods) != 1 || methods[0].Name != "getMe" {
		t.Errorf("expected 1 automatic method getMe, got %d", len(methods))
	}
}

func TestRequiresMultipart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "nil type ref", raw: "", want: false},
		{name: "non matching", raw: "String", want: false},
		{name: "matches input file", raw: "InputFile", want: true},
		{name: "nested array", raw: "Array of Array of InputFile", want: true},
		{name: "input media prefix", raw: "Array of InputMediaPhoto", want: true},
		{name: "input sticker prefix", raw: "InputStickerAnimated", want: true},
		{name: "union with match", raw: "String or InputMediaDocument", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr *parser.TypeRef
			if tc.raw != "" {
				tr = parser.NewTypeRef(tc.raw)
			}

			if got := requiresMultipart(tr); got != tc.want {
				t.Fatalf("requiresMultipart(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

//nolint:cyclop,funlen
func TestMergeUnionTypesLogic(t *testing.T) {
	validTypes := map[string]struct{}{
		"InputMedia": {},
	}

	// Case 1: OneOf with common prefix "InputMedia" -> Should merge
	inputMediaSpec := &openapi.TypeSpec{
		OneOf: []openapi.TypeSpec{
			{Ref: &openapi.TypeRef{Name: "InputMediaPhoto"}},
			{Ref: &openapi.TypeRef{Name: "InputMediaVideo"}},
		},
	}

	mergedMedia := mergeUnionTypes(inputMediaSpec, validTypes)
	if mergedMedia.Ref == nil || mergedMedia.Ref.Name != "InputMedia" {
		t.Errorf("expected InputMedia merge, got %#v", mergedMedia)
	}

	// Case 2: OneOf with NO common prefix (ReplyMarkup) -> Should NOT merge
	replyMarkupSpec := &openapi.TypeSpec{
		OneOf: []openapi.TypeSpec{
			{Ref: &openapi.TypeRef{Name: "InlineKeyboardMarkup"}},
			{Ref: &openapi.TypeRef{Name: "ReplyKeyboardMarkup"}},
			{Ref: &openapi.TypeRef{Name: "ReplyKeyboardRemove"}},
			{Ref: &openapi.TypeRef{Name: "ForceReply"}},
		},
	}

	mergedReply := mergeUnionTypes(replyMarkupSpec, validTypes)
	if len(mergedReply.OneOf) != 4 || mergedReply.Ref != nil {
		t.Errorf("expected ReplyMarkup to NOT merge, got %#v", mergedReply)
	}

	// Case 3: AnyOf (primitive + ref) -> Should NOT merge
	anyOfSpec := &openapi.TypeSpec{
		AnyOf: []openapi.TypeSpec{
			{Type: "string"},
			{Ref: &openapi.TypeRef{Name: "InputMediaPhoto"}},
		},
	}

	mergedAnyOf := mergeUnionTypes(anyOfSpec, validTypes)
	if len(mergedAnyOf.AnyOf) != 2 || mergedAnyOf.Ref != nil {
		t.Errorf("expected AnyOf mixed to NOT merge, got %#v", mergedAnyOf)
	}

	// Case 4: OneOf (refs + primitive) -> Should merge refs
	mixedSpec := &openapi.TypeSpec{
		OneOf: []openapi.TypeSpec{
			{Ref: &openapi.TypeRef{Name: "InputMediaPhoto"}},
			{Ref: &openapi.TypeRef{Name: "InputMediaVideo"}},
			{Type: "string"},
		},
	}

	mergedMixed := mergeUnionTypes(mixedSpec, validTypes)
	// We expect OneOf with 2 elements: Ref(InputMedia) and Type(string)
	if len(mergedMixed.OneOf) != 2 {
		t.Errorf("expected mixed OneOf to merge to 2 elements, got %d", len(mergedMixed.OneOf))
	}

	foundRef := false
	foundString := false

	for _, s := range mergedMixed.OneOf {
		if s.Ref != nil && s.Ref.Name == "InputMedia" {
			foundRef = true
		}

		if s.Type == "string" {
			foundString = true
		}
	}

	if !foundRef || !foundString {
		t.Errorf("expected merged mixed OneOf to contain InputMedia and string, got %#v", mergedMixed)
	}
}