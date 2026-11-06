package ui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/vivaui/viva/text"
)

func TestDeclarationOffset(t *testing.T) {
	source := "public class Hello {\n  static class Nested {}\n  record Item(String value) {}\n  enum Choice { A }\n  interface Listener {}\n}"
	for _, name := range []string{"Nested", "Item", "Choice", "Listener"} {
		offset := declarationOffset(source, name)
		if offset < 0 || !strings.HasPrefix(source[offset:], name) {
			t.Errorf("declaration %q was not found at its name", name)
		}
	}
	if declarationOffset(source, "NestedClass") >= 0 {
		t.Fatal("matched a longer identifier")
	}
	if declarationOffset(source, "Missing") >= 0 {
		t.Fatal("matched a missing declaration")
	}
}

func TestRevealDocumentSelectsDeclarationAndFocusesEditor(t *testing.T) {
	source := "public class Hello {\n  static class Nested {}\n}"
	doc := &document{source: source, editor: text.NewEditor(source)}
	m := &Model{}
	m.revealDocument(doc, "Nested")
	want := utf8.RuneCountInString(source[:strings.Index(source, "Nested")])
	selection := doc.editor.Selection()
	if doc.editor.Caret() != want || selection.Start != want || selection.End != want+len("Nested") {
		t.Fatalf("declaration selection is %+v with caret %d, want %d..%d", selection, doc.editor.Caret(), want, want+len("Nested"))
	}
	if !m.focusSource {
		t.Fatal("declaration navigation did not request editor focus")
	}
}
