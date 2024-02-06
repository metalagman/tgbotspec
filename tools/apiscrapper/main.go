package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"imposter/tools/apiscrapper/parser"
	"log"
)

func fetchSpec() ([]byte, error) {
	const fetchURL = "https://core.telegram.org/bots/api"

	client := resty.New()

	resp, err := client.R().Get(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
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
	log.Println(t)

	//v, err := parser.ParseType(doc, "update")
	//log.Printf("%#v", v)
}
