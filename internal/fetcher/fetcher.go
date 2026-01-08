package fetcher

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

var (
	fetchURL       = "https://core.telegram.org/bots/api"
	newRestyClient = func() *resty.Client { return resty.New() }
)

const (
	cacheFile     = "spec_cache.html"
	cacheLimit    = 24 * time.Hour
	cacheFilePerm = 0o644
)

// Document returns a goquery document sourced from the Telegram Bot API docs.
// It uses a cached copy when available to avoid repeated network usage.
func Document() (*goquery.Document, error) {
	html, err := HTML()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("goquery: %w", err)
	}

	return doc, nil
}

// HTML retrieves the raw HTML of the Telegram Bot API docs, leveraging a local
// cache to reduce the number of network requests.
func HTML() ([]byte, error) {
	if fileInfo, err := os.Stat(cacheFile); err == nil {
		age := time.Since(fileInfo.ModTime())
		if age < cacheLimit {
			slog.Info(
				"fetcher: using cached spec",
				"file", cacheFile,
				"age", age.Truncate(time.Second),
			)

			data, err := os.ReadFile(cacheFile)
			if err != nil {
				return nil, fmt.Errorf("read cache: %w", err)
			}

			return data, nil
		}

		slog.Info(
			"fetcher: cache expired, refetching",
			"file", cacheFile,
			"age", age.Truncate(time.Second),
			"limit", cacheLimit,
		)
	}

	client := newRestyClient()

	resp, err := client.R().Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	slog.Info(
		"fetcher: fetched spec",
		"url", fetchURL,
		"bytes", len(resp.Body()),
	)

	if err := os.WriteFile(cacheFile, resp.Body(), cacheFilePerm); err != nil {
		return nil, fmt.Errorf("write cache: %w", err)
	}

	slog.Info("fetcher: wrote cache file", "file", cacheFile)

	return resp.Body(), nil
}
