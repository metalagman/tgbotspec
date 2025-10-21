package fetcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentUsesCache(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	cachePath := filepath.Join(tmpDir, cacheFile)
	html := `<html><body><p>cached</p></body></html>`
	if err := os.WriteFile(cachePath, []byte(html), cacheFilePerm); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	doc, err := Document()
	if err != nil {
		t.Fatalf("Document returned error: %v", err)
	}

	text := strings.TrimSpace(doc.Find("p").Text())
	if text != "cached" {
		t.Fatalf("expected cache contents, got %q", text)
	}
}
