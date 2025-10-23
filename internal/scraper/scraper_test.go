package scraper //nolint:testpackage // tests rely on internal helper hooks

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"tgbotspec/internal/parser"
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
}

func TestRunWritesOpenAPISpec(t *testing.T) {
	original := fetchDocument

	t.Cleanup(func() {
		fetchDocument = original
	})

	html := `<html><head><title>Mock API</title></head><body><p><strong>Bot API 7.0</strong></p></body></html>`
	fetchDocument = func() (*goquery.Document, error) {
		return docFromString(t, html), nil
	}

	var buf bytes.Buffer
	if err := Run(&buf); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected rendered OpenAPI output")
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
