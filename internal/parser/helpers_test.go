package parser //nolint:testpackage // tests verify internal helpers

import (
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
