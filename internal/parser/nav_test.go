package parser //nolint:testpackage // tests verify internal helpers

import "testing"

func TestParseNavAndAllNavs(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<h3><a class="anchor" name="methods"></a>Methods</h3>
		<h4><a class="anchor" name="getme"></a>getMe</h4>
		<h4><a class="anchor" name="Sticker"></a>Sticker</h4>
		<h4><a class="anchor" name="double"></a>double word</h4>
		<h3><a class="anchor" name="types"></a>Types</h3>
		<h4><a class="anchor" name="user"></a>User</h4>
		</body></html>
	`)

	methods := ParseNav(doc, "methods")
	if len(methods) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(methods))
	}

	if methods[0].Mode != ParseModeMethod || methods[1].Mode != ParseModeType {
		t.Fatalf("unexpected modes: %#v", methods)
	}

	targets := ParseAllNavs(doc, []string{"methods", "types"})
	if len(targets) != 3 {
		t.Fatalf("expected 3 combined targets, got %d", len(targets))
	}
}

func TestParseNavLists(t *testing.T) {
	doc := mustDoc(t, `
		<html><body>
		<a data-target="#TypeOne">TypeOne</a>
		<a data-target="#methodOne">methodOne</a>
		<a data-target="#type-two">skip-me</a>
		<a data-target="external">External</a>
		<a data-target="#Blank">   </a>
		<a data-target="#   ">Whitespace</a>
		<a data-target="#">EmptyAnchor</a>
		<a>NoAttr</a>
		<ul class="nav navbar-nav navbar-default affix">
		  <li><a href="#Additional">Additional</a></li>
		  <li><a href="#TypeOne">Duplicate</a></li>
		  <li><a href="  ">No anchor</a></li>
		  <li><a href="#">Hash only</a></li>
		</ul>
		</body></html>
	`)

	list := ParseNavLists(doc)
	if len(list) != 3 {
		t.Fatalf("expected 3 unique targets, got %d", len(list))
	}

	modes := map[string]ParseMode{}
	for _, target := range list {
		modes[target.Name] = target.Mode
		if target.Anchor == "type-two" {
			t.Fatalf("expected anchors with dash to be skipped")
		}
	}

	if modes["TypeOne"] != ParseModeType {
		t.Fatalf("TypeOne should be ParseModeType")
	}

	if modes["methodOne"] != ParseModeMethod {
		t.Fatalf("methodOne should be ParseModeMethod")
	}
}

func TestParseNavListsIgnoresEmptyAnchors(t *testing.T) {
	doc := mustDoc(
		t,
		`<html><body>
<a data-target="#">Blank</a>
<ul class="nav navbar-nav navbar-default affix"><li><a href="">Empty</a></li></ul>
</body></html>`,
	)
	if targets := ParseNavLists(doc); len(targets) != 0 {
		t.Fatalf("expected no targets, got %#v", targets)
	}
}

func TestParseNavListsSkipsHyphenAnchors(t *testing.T) {
	doc := mustDoc(
		t,
		`<html><body><a data-target="#type-two">type-two</a></body></html>`,
	)
	if targets := ParseNavLists(doc); len(targets) != 0 {
		t.Fatalf("expected anchors with hyphen to be skipped, got %#v", targets)
	}
}

func TestIsFirstLetterCapital(t *testing.T) {
	if !isFirstLetterCapital("Type") {
		t.Fatalf("expected Type to be capitalized")
	}

	if isFirstLetterCapital("method") {
		t.Fatalf("expected method to be lowercase")
	}

	if isFirstLetterCapital("") {
		t.Fatalf("empty string should return false")
	}
}
