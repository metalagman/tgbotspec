package parser

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"unicode"
)

type TypeRef struct {
	RawType string
}

func NewTypeRef(rawType string) *TypeRef {
	return &TypeRef{RawType: rawType}
}

type TypeDef struct {
	Anchor      string
	Name        string
	Description []string
	Notes       []string
	Fields      map[string]TypeFieldDef
}

type TypeFieldDef struct {
	Name        string
	TypeRef     *TypeRef
	Required    bool
	Description string
}

type MethodDef struct {
	Anchor      string
	Name        string
	Description []string
	Notes       []string
	Params      map[string]MethodParamDef
}

type MethodParamDef struct {
	Name        string
	TypeRef     *TypeRef
	Required    bool
	Description string
}

var ErrElementNotFound = errors.New("element not found")

func ParseType(doc *goquery.Document, anchor string) (*TypeDef, error) {
	res := &TypeDef{
		Anchor: anchor,
		Fields: make(map[string]TypeFieldDef),
	}

	el := doc.Find("h4").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[Name='%s']", anchor))
	})
	// header with anchor not found
	if el.Length() == 0 {
		return nil, ErrElementNotFound
	}
	res.Name = el.Text()

	el = el.NextUntil("h4")

	el.NextFilteredUntil("p", "table").Each(func(index int, p *goquery.Selection) {
		res.Description = append(res.Description, p.Text())
	})

	el.Find("table tbody tr").Each(func(index int, tr *goquery.Selection) {
		fieldName := ""
		fieldDef := TypeFieldDef{}
		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			switch tdIndex {
			case 0:
				fieldName = td.Text()
			case 1:
				fieldDef.TypeRef = NewTypeRef(td.Text())
			case 2:
				fieldDef.Description = td.Text()
			}
		})
		fieldDef.Required = !strings.HasPrefix(fieldDef.Description, "Optional.")
		res.Fields[fieldName] = fieldDef
	})

	el.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		res.Notes = append(res.Notes, p.Text())
	})

	return res, nil
}

func ParseMethod(doc *goquery.Document, anchor string) (*MethodDef, error) {
	res := &MethodDef{
		Anchor: anchor,
		Params: make(map[string]MethodParamDef),
	}

	el := doc.Find("h4").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().First().Is(fmt.Sprintf("a.anchor[Name='%s']", anchor))
	})
	// header with anchor not found
	if el.Length() == 0 {
		return nil, ErrElementNotFound
	}
	res.Name = el.Text()

	el = el.NextUntil("h4")

	el.NextFilteredUntil("p", "table").Each(func(index int, p *goquery.Selection) {
		res.Description = append(res.Description, p.Text())
	})

	el.Find("table tbody tr").Each(func(index int, tr *goquery.Selection) {
		name := ""
		def := MethodParamDef{}
		tr.Find("td").Each(func(tdIndex int, td *goquery.Selection) {
			switch tdIndex {
			case 0:
				name = td.Text()
			case 1:
				def.TypeRef = NewTypeRef(td.Text())
			case 2:
				def.Required = td.Text() == "Yes"
			case 3:
				def.Description = td.Text()
			}
		})

		res.Params[name] = def
	})

	el.Find("blockquote p").Each(func(index int, p *goquery.Selection) {
		res.Notes = append(res.Notes, p.Text())
	})

	return res, nil
}

type ParseMode int

const (
	ParseModeType ParseMode = iota
	ParseModeMethod
)

type ParseTarget struct {
	Anchor string
	Name   string
	Mode   ParseMode
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
		return s.Children().First().Is(fmt.Sprintf("a.anchor[Name='%s']", anchor))
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
				Anchor: id,
				Name:   name,
			}
			if isFirstLetterCapital(t.Name) {
				t.Mode = ParseModeType
			} else {
				t.Mode = ParseModeMethod
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
