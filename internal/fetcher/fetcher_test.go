package fetcher //nolint:testpackage // coverage test exercises internal cache behavior

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
)

func useTempWorkDir(t *testing.T) string {
	t.Helper()

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

	return tmpDir
}

func TestDocumentUsesCache(t *testing.T) {
	tmpDir := useTempWorkDir(t)

	cachePath := filepath.Join(tmpDir, "spec_cache.html")

	html := `<html><body><p>cached</p></body></html>`
	if err := os.WriteFile(cachePath, []byte(html), 0o644); err != nil {
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

func TestHTMLFetchesAndCaches(t *testing.T) {
	tmpDir := useTempWorkDir(t)

	origURL := fetchURL
	origNewClient := newRestyClient

	fetchURL = "https://example.com/bots/api"

	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	t.Cleanup(httpmock.DeactivateAndReset)

	newRestyClient = func() *resty.Client {
		return client
	}

	t.Cleanup(func() {
		fetchURL = origURL
		newRestyClient = origNewClient
	})

	const body = `<html><body><p>fresh</p></body></html>`
	httpmock.RegisterResponder("GET", fetchURL, httpmock.NewStringResponder(200, body))

	data, err := HTML()
	if err != nil {
		t.Fatalf("HTML: %v", err)
	}

	if string(data) != body {
		t.Fatalf("expected fetched body %q, got %q", body, string(data))
	}

	info := httpmock.GetCallCountInfo()
	if info["GET "+fetchURL] != 1 {
		t.Fatalf("expected one fetch call, got counts %+v", info)
	}

	cached, err := os.ReadFile(filepath.Join(tmpDir, cacheFile))
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}

	if string(cached) != body {
		t.Fatalf("expected cache contents %q, got %q", body, string(cached))
	}

	httpmock.ZeroCallCounters()

	data, err = HTML()
	if err != nil {
		t.Fatalf("HTML second call: %v", err)
	}

	if string(data) != body {
		t.Fatalf("expected cached body %q on second call, got %q", body, string(data))
	}

	if total := httpmock.GetTotalCallCount(); total != 0 {
		t.Fatalf("expected cache read without network call, got %d calls", total)
	}
}

func TestHTMLFetchError(t *testing.T) {
	useTempWorkDir(t)

	origURL := fetchURL
	origNewClient := newRestyClient

	fetchURL = "https://example.com/fail"

	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	t.Cleanup(httpmock.DeactivateAndReset)

	newRestyClient = func() *resty.Client {
		return client
	}

	t.Cleanup(func() {
		fetchURL = origURL
		newRestyClient = origNewClient
	})

	httpmock.RegisterResponder("GET", fetchURL, httpmock.NewErrorResponder(errors.New("network down")))

	if _, err := HTML(); err == nil {
		t.Fatal("expected HTML to return error on fetch failure")
	}

	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Fatalf("expected no cache file on fetch failure, stat err = %v", err)
	}
}
