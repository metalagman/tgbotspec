package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"imposter/tools/apiscrapper/parser"
	"log"
	"os"
	"text/template"
	"time"
)

const (
	fetchURL   = "https://core.telegram.org/bots/api"
	cacheFile  = "spec_cache.html" // Changed extension to .html
	cacheLimit = 24 * time.Hour    // Cache validity duration, e.g., 24 hours
)

// fetchSpec tries to load the API specification from a cache file;
// if the cache is outdated or does not exist, it fetches it from the internet.
func fetchSpec() ([]byte, error) {
	// Check if the cache file exists and is valid
	if fileInfo, err := os.Stat(cacheFile); err == nil {
		if time.Since(fileInfo.ModTime()) < cacheLimit {
			// Cache is valid, load the data from the file
			data, err := os.ReadFile(cacheFile)
			if err != nil {
				return nil, fmt.Errorf("read cache: %w", err)
			}
			return data, nil
		}
	}

	// Cache is invalid or does not exist, fetch from the internet
	client := resty.New()
	resp, err := client.R().Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	// Save the fetched data to the cache file for future requests
	if err := os.WriteFile(cacheFile, resp.Body(), 0644); err != nil {
		return nil, fmt.Errorf("write cache: %w", err)
	}

	return resp.Body(), nil
}

func getDoc() (*goquery.Document, error) {
	html, err := fetchSpec()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("goquery: %w", err)
	}

	return doc, nil
}

func main() {
	doc, err := getDoc()
	if err != nil {
		log.Fatalf(err.Error())
	}

	sections := []string{
		"getting-updates",
		"available-types",
		"available-methods",
		"updating-messages",
		"stickers",
		"inline-mode",
		"payments",
		"telegram-passport",
		"games",
	}

	t := parser.ParseAllNavs(doc, sections)

	params := struct {
		Types []parser.TypeDef
	}{
		Types: make([]parser.TypeDef, 0),
	}

	for _, v := range t {
		switch v.Mode {
		//case parser.ParseModeMethod:
		//	md, _ := parser.ParseMethod(doc, v.Anchor)
		case parser.ParseModeType:
			td, err := parser.ParseType(doc, v.Anchor)
			if err != nil {
				log.Println(err)
				continue
			}
			params.Types = append(params.Types, *td)
		}
	}

	renderTemplate(params)

	//v, err := parser.ParseType(doc, "update")
	//log.Printf("%#v", v)
}

func renderTemplate(params interface{}) {
	// Path to your template file
	templateFilePath := "openapi.yaml.gotmpl"

	// Load and parse the template file
	tmpl, err := template.ParseFiles(templateFilePath)
	if err != nil {
		log.Fatalf("Error loading template file: %v", err)
	}

	// Execute the template with your data
	err = tmpl.Execute(os.Stdout, params)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
}
