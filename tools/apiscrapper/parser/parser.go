package parser

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"unicode"
)

type TypeRef struct {
	rawType string
}

func NewTypeRef(rawType string) *TypeRef {
	return &TypeRef{rawType: rawType}
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
	typeRef     *TypeRef
	required    bool
	description string
}

type MethodDef struct {
	anchor      string
	name        string
	description []string
	notes       []string
	params      map[string]MethodParamDef
}

type MethodParamDef struct {
	name        string
	typeRef     *TypeRef
	required    bool
	description string
}

var ErrElementNotFound = errors.New("element not found")

func ParseType(doc *goquery.Document, anchor string) (*TypeDef, error) {
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
				fieldDef.typeRef = NewTypeRef(td.Text())
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

func ParseMethod(doc *goquery.Document, anchor string) (*MethodDef, error) {
	res := &MethodDef{
		anchor: anchor,
		params: make(map[string]MethodParamDef),
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
		name := ""
		def := MethodParamDef{}
		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			log.Println(tdIndex, td.Text())
			switch tdIndex {
			case 0:
				name = td.Text()
			case 1:
				def.typeRef = NewTypeRef(td.Text())
			case 2:
				def.required = td.Text() == "Yes"
			case 3:
				def.description = td.Text()
			}
			log.Println(def)
		})

		res.params[name] = def
	})

	el.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		res.notes = append(res.notes, p.Text())
	})

	return res, nil
}

type ParseMode int

const (
	ParseModeType ParseMode = iota
	ParseModeMethod
)

type ParseTarget struct {
	anchor string
	name   string
	mode   ParseMode
}

func isFirstLetterCapital(str string) bool {
	for _, r := range str {
		return unicode.IsUpper(r)
	}
	return false
}

func containsExactlyOneWord(s string) bool {
	words := strings.Fields(s)
	return len(words) == 1
}

func ParseNav(doc *goquery.Document, anchor string) []ParseTarget {
	var res []ParseTarget

	el := doc.Find("h3").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[name='%s']", anchor))
	})

	el = el.NextFilteredUntil("h4", "h3").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is("a.anchor")
	})

	el.Each(func(i int, s *goquery.Selection) {
		name := s.Text()
		if !containsExactlyOneWord(name) {
			return
		}
		s = s.Find("a.anchor")
		if id, ok := s.Attr("name"); ok {
			t := ParseTarget{
				anchor: id,
				name:   name,
			}
			if isFirstLetterCapital(t.name) {
				t.mode = ParseModeType
			} else {
				t.mode = ParseModeMethod
			}
			res = append(res, t)
		}
	})

	return res
}

func ParseAllNavs(doc *goquery.Document, anchors []string) []ParseTarget {
	var targets []ParseTarget

	for _, s := range anchors {
		t := ParseNav(doc, s)
		targets = append(targets, t...)
	}

	return targets
}
