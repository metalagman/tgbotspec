package fetcher

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

const (
	fetchURL   = "https://core.telegram.org/bots/api"
	cacheFile  = "spec_cache.html"
	cacheLimit = 24 * time.Hour
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
			log.Printf("fetcher: using cached spec %s (age %s)", cacheFile, age.Truncate(time.Second))
			data, err := os.ReadFile(cacheFile)
			if err != nil {
				return nil, fmt.Errorf("read cache: %w", err)
			}
			return data, nil
		}
		log.Printf("fetcher: cache expired for %s (age %s > %s), refetching", cacheFile, age.Truncate(time.Second), cacheLimit)
	}

	client := resty.New()
	resp, err := client.R().Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	log.Printf("fetcher: fetched spec from %s (%d bytes)", fetchURL, len(resp.Body()))

	if err := os.WriteFile(cacheFile, resp.Body(), 0644); err != nil {
		return nil, fmt.Errorf("write cache: %w", err)
	}
	log.Printf("fetcher: wrote cache file %s", cacheFile)

	return resp.Body(), nil
}
