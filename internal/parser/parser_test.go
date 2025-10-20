package parser

import (
	"testing"
)

func TestExtractReturnType(t *testing.T) {
	cases := []struct {
		name     string
		paras    []string
		expected string
	}{
		{
			name:     "Returns an Array of Update objects",
			paras:    []string{"Use this method to receive incoming updates using long polling. Returns an Array of Update objects."},
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
			name:     "On success, the sent Message is returned",
			paras:    []string{"Use this method to send text messages. On success, the sent Message is returned."},
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
			name:     "On success array of MessageId",
			paras:    []string{"On success, an array of MessageId of the sent messages is returned."},
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
			name:     "Conditional return message or true",
			paras:    []string{"On success, if the message is not an inline message, the Message is returned, otherwise True is returned."},
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

func TestReturnTypeToSpec(t *testing.T) {
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
