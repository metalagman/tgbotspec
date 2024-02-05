package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
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

type TypeDef struct {
	anchor      string
	name        string
	description []string
	notes       []string
	fields      map[string]TypeFieldDef
}

type TypeFieldDef struct {
	name        string
	typeRef     string
	description string
	required    bool
}

var ErrElementNotFound = errors.New("element not found")

func parseType(doc *goquery.Document, anchor string) (*TypeDef, error) {
	res := &TypeDef{
		anchor: anchor,
		fields: make(map[string]TypeFieldDef),
	}

	el := doc.Find("h4").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[name='%s']", anchor))
	})
	// header with anchor not found
	if el.Length() == 0 {
		return nil, ErrElementNotFound
	}
	res.name = el.Text()

	el = el.NextUntil("h4")

	el.NextFilteredUntil("p", "table").Each(func(index int, p *goquery.Selection) {
		res.description = append(res.description, p.Text())
	})

	el.Find("table tbody tr").Each(func(index int, tr *goquery.Selection) {
		fieldName := ""
		fieldDef := TypeFieldDef{}
		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			log.Println(tdIndex, td.Text())
			switch tdIndex {
			case 0:
				fieldName = td.Text()
			case 1:
				fieldDef.typeRef = td.Text()
			case 2:
				fieldDef.description = td.Text()
			}
			log.Println(fieldDef)
		})
		fieldDef.required = !strings.HasPrefix(fieldDef.description, "Optional.")
		res.fields[fieldName] = fieldDef
	})

	el.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		res.notes = append(res.notes, p.Text())
	})

	return res, nil
}

func main() {
	doc, err := getDoc()
	if err != nil {
		log.Fatalf(err.Error())
	}

	v, err := parseType(doc, "update")
	log.Printf("%#v", v)
}
