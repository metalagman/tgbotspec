package parser //nolint:testpackage // tests verify internal helpers

import (
	"errors"
	"reflect"
	"testing"
)

//nolint:cyclop,funlen // comprehensive table checks parsing branches
func TestParseMethodSuccess(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h3>Messaging API</h3>
		<h4><a class="anchor" name="testmethod"></a>testMethod</h4>
		<p>Use this method to do something useful. Returns the created invoice link as String on success.</p>
		<p>Second line of the description.</p>
		<table class="table">
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
			<tr>
			  <td>from_chat_id</td>
			  <td>Integer or String</td>
			  <td>Optional</td>
			  <td>Optional. Source chat identifier</td>
			</tr>
		  </tbody>
		</table>
		<blockquote><p>Only works in supergroups.</p></blockquote>
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

	if got, ok := method.Params["from_chat_id"]; !ok || got.TypeRef.RawType != "Integer" || got.Required {
		t.Fatalf("from_chat_id param not normalized: %#v", got)
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

func TestNormalizeReturnTypePhrase(t *testing.T) { //nolint:cyclop // broad phrase coverage requires many cases
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

//nolint:cyclop
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

	if got := extractReturnTypeFromSentence("On success returns Message."); got != "Message" {
		t.Fatalf("expected Message from 'On success returns Message.', got %q", got)
	}

	if got := extractReturnTypeFromSentence("On success, returns Message."); got != "Message" {
		t.Fatalf("expected Message from 'On success, returns Message.', got %q", got)
	}

	if extractReturnTypeFromSentence("On success, returns unknown data.") != "" {
		t.Fatalf("expected empty for unknown data after returns")
	}

	if got := extractReturnTypeFromSentence("The Message is returned."); got != "Message" {
		t.Fatalf("expected Message from final clause, got %q", got)
	}
}

func TestExtractReturnType(t *testing.T) { //nolint:funlen // large table ensures coverage of phrasing variants
	cases := []struct {
		name     string
		paras    []string
		expected string
	}{
		{
			name: "Returns an Array of Update objects",
			paras: []string{
				"Use this method to receive incoming updates using long polling. " +
					"Returns an Array of Update objects.",
			},
			expected: "Array of Update",
		},
		{
			name:     "Returns True on success",
			paras:    []string{"Returns True on success."},
			expected: "True",
		},
		{
			name:     "Skips blank paragraphs",
			paras:    []string{"   ", "Returns Boolean."},
			expected: "Boolean",
		},
		{
			name:     "On success, returns a WebhookInfo object",
			paras:    []string{"On success, returns a WebhookInfo object."},
			expected: "WebhookInfo",
		},
		{
			name:     "Returns StarAmount on success",
			paras:    []string{"Returns StarAmount on success."},
			expected: "StarAmount",
		},
		{
			name:     "Returns StickerSet without qualifier",
			paras:    []string{"Returns StickerSet."},
			expected: "StickerSet",
		},
		{
			name: "On success, the sent Message is returned",
			paras: []string{
				"Use this method to send text messages. On success, the sent Message is returned.",
			},
			expected: "Message",
		},
		{
			name:     "Returns the MessageId of the sent message",
			paras:    []string{"Returns the MessageId of the sent message on success."},
			expected: "MessageId",
		},
		{
			name:     "No return info",
			paras:    []string{"Use this method to configure webhook parameters."},
			expected: "",
		},
		{
			name: "On success array of MessageId",
			paras: []string{
				"On success, an array of MessageId of the sent messages is returned.",
			},
			expected: "Array of MessageId",
		},
		{
			name:     "On success returns array",
			paras:    []string{"On success, returns Array of Sticker objects."},
			expected: "Array of Sticker",
		},
		{
			name:     "NBSP between success and returns",
			paras:    []string{"On success\u00a0returns Message."},
			expected: "Message",
		},
		{
			name:     "Returns link as String",
			paras:    []string{"Returns the created invoice link as String on success."},
			expected: "String",
		},
		{
			name:     "Returns user info in form of object",
			paras:    []string{"Returns basic information about the bot in form of a User object."},
			expected: "User",
		},
		{
			name:     "Returns information about created topic",
			paras:    []string{"Returns information about the created topic as a ForumTopic object."},
			expected: "ForumTopic",
		},
		{
			name: "Conditional return message or true",
			paras: []string{
				"On success, if the message is not an inline message, the Message is returned, otherwise True is returned.",
			},
			expected: "Message or True",
		},
		{
			name:     "On success stopped poll returned",
			paras:    []string{"On success, the stopped Poll is returned."},
			expected: "Poll",
		},
		{
			name:     "On success returns array no comma",
			paras:    []string{"On success returns Array of Sticker."},
			expected: "Array of Sticker",
		},
		{
			name:     "Returns Poll upon success",
			paras:    []string{"Returns Poll upon success."},
			expected: "Poll",
		},
		{
			name:     "Returns Message when condition",
			paras:    []string{"Returns Message when the user is online."},
			expected: "Message",
		},
		{
			name:     "Message are returned",
			paras:    []string{"On success, Messages are returned."},
			expected: "Messages",
		},
		{
			name:     "Returns Message with parentheses",
			paras:    []string{"Returns Message (see docs) on success."},
			expected: "Message",
		},
		{
			name:     "Returns info in the form of",
			paras:    []string{"On success, returns information about the chat in the form of an ChatFullInfo object."},
			expected: "ChatFullInfo",
		},
	}

	for _, tc := range cases {
		got := extractReturnType(tc.paras)
		if got != tc.expected {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.expected, got)
		}
	}
}
