package parser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type ParseMode int

const (
	ParseModeType ParseMode = iota
	ParseModeMethod
)

// ParseTarget describes an extracted navigation anchor and whether it
// represents a method or a type section.
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

// ParseNav scans the documentation for a navigation section identified by the
// given anchor and returns the immediate method/type targets below it.
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

// ParseAllNavs aggregates ParseNav results for a list of anchors.
func ParseAllNavs(doc *goquery.Document, anchors []string) []ParseTarget {
	var targets []ParseTarget

	for _, s := range anchors {
		t := ParseNav(doc, s)
		targets = append(targets, t...)
	}

	return targets
}

// ParseNavLists extracts targets from the navigation lists rendered at the top
// of the Telegram docs, deduplicating anchors and inferring their mode.
func ParseNavLists(doc *goquery.Document) []ParseTarget {
	var res []ParseTarget
	seen := make(map[string]struct{})

	addTarget := func(anchor, name string) {
		anchor = strings.TrimSpace(anchor)
		if anchor == "" || strings.Contains(anchor, "-") {
			return
		}
		if _, exists := seen[anchor]; exists {
			return
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		t := ParseTarget{
			Anchor: anchor,
			Name:   name,
		}
		if isFirstLetterCapital(name) {
			t.Mode = ParseModeType
		} else {
			t.Mode = ParseModeMethod
		}
		res = append(res, t)
		seen[anchor] = struct{}{}
	}

	doc.Find("a[data-target]").Each(func(i int, s *goquery.Selection) {
		target, _ := s.Attr("data-target")
		target = strings.TrimSpace(target)
		if !strings.HasPrefix(target, "#") {
			return
		}
		anchor := strings.TrimPrefix(target, "#")
		if anchor == "" {
			return
		}
		addTarget(anchor, s.Text())
	})

	doc.Find("ul.nav.navbar-nav.navbar-default.affix a[href^='#']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		anchor := strings.TrimPrefix(strings.TrimSpace(href), "#")
		if anchor == "" {
			return
		}
		addTarget(anchor, s.Text())
	})

	return res
}
