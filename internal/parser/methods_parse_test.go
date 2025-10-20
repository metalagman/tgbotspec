package parser

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func mustDoc(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to build document: %v", err)
	}
	return doc
}

func TestParseMethodSuccess(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h3>Messaging API</h3>
		<h4>
		  <a class="anchor" name="testmethod"></a>testMethod
		  <div>
		    <table>
		      <tbody>
		        <tr>
		          <td>chat_id</td>
		          <td>String</td>
		          <td>Optional field</td>
		          <td>Optional. Chat identifier</td>
		        </tr>
		        <tr>
		          <td>limit</td>
		          <td>Integer</td>
		          <td>Maximum number of items to return</td>
		          <td>Limit description</td>
		        </tr>
		      </tbody>
		    </table>
		    <blockquote><p>Only works in supergroups.</p></blockquote>
		  </div>
		</h4>
		<table class="table"><tbody><tr><td>ignored</td></tr></tbody></table>
		<p>Use this method to do something useful. Returns the created invoice link as String on success.</p>
		<p>Second line of the description.</p>
		<h4><a class="anchor" name="next"></a>next</h4>
		</body></html>
	`)

	method, err := ParseMethod(doc, "testmethod")
	if err != nil {
		t.Fatalf("ParseMethod returned error: %v", err)
	}
	if method.Anchor != "testmethod" {
		t.Fatalf("unexpected anchor %q", method.Anchor)
	}
	if len(method.Tags) != 1 || method.Tags[0] != "Messaging API" {
		t.Fatalf("expected tag from preceding h3, got %#v", method.Tags)
	}
	if got := len(method.Description); got != 2 {
		t.Fatalf("expected two description paragraphs, got %d", got)
	}
	if method.Return == nil || method.Return.RawType != "String" {
		t.Fatalf("expected return type String, got %#v", method.Return)
	}
	if got, ok := method.Params["chat_id"]; !ok || got.TypeRef.RawType != "Integer" || got.Required {
		t.Fatalf("chat_id param not normalized: %#v", got)
	}
	if got, ok := method.Params["limit"]; !ok || got.TypeRef.RawType != "Integer" || !got.Required {
		t.Fatalf("limit param missing required flag: %#v", got)
	}
	if !reflect.DeepEqual(method.Notes, []string{"Only works in supergroups."}) {
		t.Fatalf("unexpected notes: %#v", method.Notes)
	}
}

func TestParseMethodErrors(t *testing.T) {
	t.Run("missing anchor", func(t *testing.T) {
		doc := mustDoc(t, `<html><body><h4><a class="anchor" name="other"></a>dummy</h4></body></html>`)
		_, err := ParseMethod(doc, "absent")
		if !errors.Is(err, ErrElementNotFound) {
			t.Fatalf("expected ErrElementNotFound, got %v", err)
		}
	})

	t.Run("return not parsed", func(t *testing.T) {
		doc := mustDoc(t, `
		<html><body>
		<h3>Section</h3>
		<h4><a class="anchor" name="noreturn"></a>noReturn</h4>
		<p>This paragraph lacks any return type hint.</p>
		<h4>next</h4>
		</body></html>`)
		_, err := ParseMethod(doc, "noreturn")
		if !errors.Is(err, ErrReturnTypeNotParsed) {
			t.Fatalf("expected ErrReturnTypeNotParsed, got %v", err)
		}
	})
}

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
	if got := len(typeDef.Fields); got != 2 {
		t.Fatalf("expected two fields, got %d", got)
	}
	fields := typeDef.Fields
	if fields[0].Name != "chat_id" || fields[0].TypeRef.RawType != "Integer" || fields[0].Required {
		t.Fatalf("chat_id field not normalized: %#v", fields[0])
	}
	if !fields[1].Required {
		t.Fatalf("expected second field to be required")
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

func TestParseNavAndAllNavs(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h3><a class="anchor" name="methods"></a>Methods</h3>
		<h4><a class="anchor" name="getme"></a>getMe</h4>
		<h4><a class="anchor" name="Sticker"></a>Sticker</h4>
		<h4><a class="anchor" name="double"></a>double word</h4>
		<h3><a class="anchor" name="types"></a>Types</h3>
		<h4><a class="anchor" name="user"></a>User</h4>
		</body></html>
	`)

	methods := ParseNav(doc, "methods")
	if len(methods) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(methods))
	}
	if methods[0].Mode != ParseModeMethod || methods[1].Mode != ParseModeType {
		t.Fatalf("unexpected modes: %#v", methods)
	}

	targets := ParseAllNavs(doc, []string{"methods", "types"})
	if len(targets) != 3 {
		t.Fatalf("expected 3 combined targets, got %d", len(targets))
	}
}

func TestParseNavLists(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<a data-target="#TypeOne">TypeOne</a>
		<a data-target="#methodOne">methodOne</a>
		<a data-target="#type-two">skip-me</a>
		<a data-target="external">External</a>
		<a data-target="#Blank">   </a>
		<a data-target="#   ">Whitespace</a>
		<a data-target="#">EmptyAnchor</a>
		<a>NoAttr</a>
		<ul class="nav navbar-nav navbar-default affix">
		  <li><a href="#Additional">Additional</a></li>
		  <li><a href="#TypeOne">Duplicate</a></li>
		  <li><a href="  ">No anchor</a></li>
		  <li><a href="#">Hash only</a></li>
		</ul>
		</body></html>
	`)

	list := ParseNavLists(doc)
	if len(list) != 3 {
		t.Fatalf("expected 3 unique targets, got %d", len(list))
	}
	modes := map[string]ParseMode{}
	for _, target := range list {
		modes[target.Name] = target.Mode
		if target.Anchor == "type-two" {
			t.Fatalf("expected anchors with dash to be skipped")
		}
	}
	if modes["TypeOne"] != ParseModeType {
		t.Fatalf("TypeOne should be ParseModeType")
	}
	if modes["methodOne"] != ParseModeMethod {
		t.Fatalf("methodOne should be ParseModeMethod")
	}
}

func TestParseNavListsIgnoresEmptyAnchors(t *testing.T) {
	doc := mustDoc(t, `<html><body><a data-target="#">Blank</a><ul class="nav navbar-nav navbar-default affix"><li><a href="">Empty</a></li></ul></body></html>`)
	if targets := ParseNavLists(doc); len(targets) != 0 {
		t.Fatalf("expected no targets, got %#v", targets)
	}
}

func TestParseNavListsSkipsHyphenAnchors(t *testing.T) {
	doc := mustDoc(t, `<html><body><a data-target="#type-two">type-two</a></body></html>`)
	if targets := ParseNavLists(doc); len(targets) != 0 {
		t.Fatalf("expected anchors with hyphen to be skipped, got %#v", targets)
	}
}

func TestSplitIntoSentences(t *testing.T) {
	got := splitIntoSentences("First. Second! Third")
	if len(got) != 3 {
		t.Fatalf("expected three sentences, got %#v", got)
	}
	if splitIntoSentences("   ") != nil {
		t.Fatalf("expected nil for blank input")
	}
}

func TestLooksLikeReturnType(t *testing.T) {
	if looksLikeReturnType("") {
		t.Fatalf("empty string should not look like a return type")
	}
	if !looksLikeReturnType("string") {
		t.Fatalf("primitive string should be recognized")
	}
	if looksLikeReturnType("customtype") {
		t.Fatalf("lowercase custom type should not be recognized")
	}
	if !looksLikeReturnType("Array of Message") {
		t.Fatalf("array phrasing should be recognized")
	}
}

func TestNormalizeReturnTypePhrase(t *testing.T) {
	if got := normalizeReturnTypePhrase(""); got != "" {
		t.Fatalf("expected empty string to remain empty, got %q", got)
	}
	if got := normalizeReturnTypePhrase("the created invoice link as String"); got != "String" {
		t.Fatalf("unexpected normalization result: %q", got)
	}
	if got := normalizeReturnTypePhrase("information about the chat in form of an User object"); got != "User" {
		t.Fatalf("expected User, got %q", got)
	}
	if got := normalizeReturnTypePhrase("Array of stickers"); got != "Array of stickers" {
		t.Fatalf("expected Array of stickers, got %q", got)
	}
	if got := normalizeReturnTypePhrase("information about the data in the form of a Chat object"); got != "Chat" {
		t.Fatalf("expected Chat, got %q", got)
	}
	if got := normalizeReturnTypePhrase("information about the session in the form of an User object"); got != "User" {
		t.Fatalf("expected User via 'the form of an', got %q", got)
	}
	if got := normalizeReturnTypePhrase("information about the user as an ChatMember object"); got != "ChatMember" {
		t.Fatalf("expected ChatMember via 'as an', got %q", got)
	}
	if got := normalizeReturnTypePhrase("information about the topic as a ForumTopic object"); got != "ForumTopic" {
		t.Fatalf("expected ForumTopic, got %q", got)
	}
	if got := normalizeReturnTypePhrase("the revoked invite link as ChatInviteLink"); got != "ChatInviteLink" {
		t.Fatalf("expected ChatInviteLink, got %q", got)
	}
	if got := normalizeReturnTypePhrase("Message objects that were sent"); got != "Message" {
		t.Fatalf("expected Message, got %q", got)
	}
	if got := normalizeReturnTypePhrase("an array of Message"); got != "Array of Message" {
		t.Fatalf("expected Array of Message, got %q", got)
	}
	if got := normalizeReturnTypePhrase("chat link as ChatInviteLink"); got != "ChatInviteLink" {
		t.Fatalf("expected ChatInviteLink via suffix check, got %q", got)
	}
	if got := normalizeReturnTypePhrase("integer"); got != "Integer" {
		t.Fatalf("expected Integer capitalization, got %q", got)
	}
	if got := normalizeReturnTypePhrase("string"); got != "String" {
		t.Fatalf("expected String capitalization, got %q", got)
	}
}

func TestParseReturnsClauseMarkers(t *testing.T) {
	if got := parseReturnsClause("Message on success, if needed"); got != "Message" {
		t.Fatalf("expected Message, got %q", got)
	}
	if got := parseReturnsClause("Message when ready"); got != "Message" {
		t.Fatalf("expected Message, got %q", got)
	}
	if got := parseReturnsClause("Message (deprecated) on success"); got != "Message" {
		t.Fatalf("parenthetical stop failed: %q", got)
	}
	if got := parseReturnsClause("Message [note] on success"); got != "Message" {
		t.Fatalf("bracket stop failed: %q", got)
	}
	if got := parseReturnsClause("Message, if available"); got != "Message" {
		t.Fatalf("comma stop failed: %q", got)
	}
	if got := parseReturnsClause("Message, additional context"); got != "Message" {
		t.Fatalf("generic comma stop failed: %q", got)
	}
	if got := parseReturnsClause("unknown value"); got != "" {
		t.Fatalf("unexpected detection for unknown value: %q", got)
	}
	if got := parseReturnsClause("Message; fallback"); got != "Message" {
		t.Fatalf("semicolon stop failed: %q", got)
	}
	if parseReturnsClause("") != "" {
		t.Fatalf("expected empty result for blank remainder")
	}
}

func TestParseIsReturnedClauseVariants(t *testing.T) {
	if got := parseIsReturnedClause("Array of Sticker are returned."); got != "Array of Sticker" {
		t.Fatalf("expected Array of Sticker, got %q", got)
	}
	if got := parseIsReturnedClause("Message is returned, otherwise returns True."); got != "Message or True" {
		t.Fatalf("expected Message or True, got %q", got)
	}
	if got := parseIsReturnedClause("Message is returned, or returns True."); got != "Message or True" {
		t.Fatalf("expected Message or True with or clause, got %q", got)
	}
	if parseIsReturnedClause("No return info") != "" {
		t.Fatalf("expected empty string when no return clause present")
	}
	if parseIsReturnedClause(" is returned") != "" {
		t.Fatalf("expected empty string for blank prefix")
	}
	if got := parseIsReturnedClause("Message is returned, or Message is returned."); got != "Message" {
		t.Fatalf("expected deduplicated Message, got %q", got)
	}
	if parseIsReturnedClause("data is returned.") != "" {
		t.Fatalf("expected empty result for lowercase type")
	}
}

func TestIsFirstLetterCapital(t *testing.T) {
	if !isFirstLetterCapital("Type") {
		t.Fatalf("expected Type to be capitalized")
	}
	if isFirstLetterCapital("method") {
		t.Fatalf("expected method to be lowercase")
	}
	if isFirstLetterCapital("") {
		t.Fatalf("empty string should return false")
	}
}

func TestExtractReturnTypeFromSentenceHelpers(t *testing.T) {
	if extractReturnTypeFromSentence("") != "" {
		t.Fatalf("expected empty result for blank sentence")
	}
	if extractReturnTypeFromSentence("On success, ") != "" {
		t.Fatalf("expected empty result for incomplete sentence")
	}
	if extractReturnTypeFromSentence("Status updated") != "" {
		t.Fatalf("expected empty result when sentence lacks return keyword")
	}
	if extractReturnTypeFromSentence("On success, returns data if available.") != "" {
		t.Fatalf("expected empty result for unrecognized remainder")
	}
	if extractReturnTypeFromSentence("On success returns data if available.") != "" {
		t.Fatalf("expected empty result for 'on success returns' variant")
	}
	if extractReturnTypeFromSentence("On success, Returns data if available.") != "" {
		t.Fatalf("expected case-insensitive handling")
	}
	if got := extractReturnTypeFromSentence("The Message is returned."); got != "Message" {
		t.Fatalf("expected Message from final clause, got %q", got)
	}
}

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

	if spec := NewTypeRef("Array of Message").ToTypeSpec(); spec.Type != "array" || spec.Items == nil || spec.Items.Ref == nil || spec.Items.Ref.Name != "Message" {
		t.Fatalf("array spec unexpected: %#v", spec)
	}
	if spec := NewTypeRef("Array of Array of Sticker").ToTypeSpec(); spec.Type != "array" || spec.Items == nil || spec.Items.Type != "array" {
		t.Fatalf("nested array spec unexpected: %#v", spec)
	}
	if spec := NewTypeRef("array of array of Sticker").ToTypeSpec(); spec.Type != "array" {
		t.Fatalf("expected array type for lowercase phrasing, got %#v", spec)
	}
	if spec := NewTypeRef("Integer").ToTypeSpec(); spec.Type != "integer" {
		t.Fatalf("primitive spec unexpected: %#v", spec)
	}
	if spec := NewTypeRef("int64").ToTypeSpec(); spec.Format != "int64" {
		t.Fatalf("expected int64 format, got %#v", spec)
	}
	if spec := NewTypeRef("string").ToTypeSpec(); spec.Type != "string" {
		t.Fatalf("expected string type, got %#v", spec)
	}
	if spec := NewTypeRef("True").ToTypeSpec(); spec.Default != true {
		t.Fatalf("expected True default, got %#v", spec)
	}
	if spec := NewTypeRef("Sticker or Photo").ToTypeSpec(); len(spec.AnyOf) != 2 {
		t.Fatalf("expected anyOf for union, got %#v", spec)
	}
	if spec := NewTypeRef("CustomType").ToTypeSpec(); spec.Ref == nil || spec.Ref.Name != "CustomType" {
		t.Fatalf("expected ref for custom type, got %#v", spec)
	}
	if spec := (*TypeRef)(nil).ToTypeSpec(); spec.Type != "" || spec.Ref != nil {
		t.Fatalf("expected empty spec for nil TypeRef")
	}
}
