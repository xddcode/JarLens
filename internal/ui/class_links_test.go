package ui

import (
	"testing"

	"github.com/xddcode/jarlens/internal/archive"
)

func TestServiceClassLinks(t *testing.T) {
	index := newClassLinkIndex([]archive.Entry{{ID: 4, Path: "org/example/Service.class", Name: "Service.class", Kind: archive.EntryClass}})
	source := "# comment\n  org.example.Service  \nmissing.Provider\n"
	links := buildClassLinks(source, "META-INF/services/example.Service", index)
	if len(links) != 1 || links[0].entryID != 4 {
		t.Fatalf("service links = %+v", links)
	}
	link := links[0]
	if got := source[link.start:link.end]; got != "org.example.Service" {
		t.Fatalf("service link text = %q", got)
	}
	if _, ok := classLinkAt(links, link.start+4); !ok {
		t.Fatal("service link did not contain its text offset")
	}
}

func TestJavaClassLinksIncludeNestedAndSkipCommentsAndStrings(t *testing.T) {
	index := newClassLinkIndex([]archive.Entry{
		{ID: 0, Path: "example/Outer.class", Name: "Outer.class", Kind: archive.EntryClass},
		{ID: 1, Path: "example/Outer$Inner.class", Name: "Outer$Inner.class", Kind: archive.EntryClass},
	})
	source := `class Example extends Outer {
  String text = "Inner"; // Inner
  Inner inner;
}`
	links := buildClassLinks(source, "example/Example.class", index)
	if len(links) != 2 || links[0].entryID != 0 || links[1].entryID != 1 {
		t.Fatalf("Java links = %+v", links)
	}
	if got := source[links[1].start:links[1].end]; got != "Inner" {
		t.Fatalf("nested class link text = %q", got)
	}
}

func TestJavaClassLinksIgnoreUnknownNames(t *testing.T) {
	index := newClassLinkIndex(nil)
	if links := buildClassLinks("String value = missing.Name;", "Example.class", index); len(links) != 0 {
		t.Fatalf("unknown names became links: %+v", links)
	}
}
