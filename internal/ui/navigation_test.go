package ui

import (
	"strings"
	"testing"
	"time"
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

func TestNestedDeclarationName(t *testing.T) {
	for name, want := range map[string]string{
		"Hello.class":         "Hello",
		"Hello$Nested.class":  "Nested",
		"Hello$1.class":       "",
		"Hello$1Local.class":  "Local",
		"Hello$12Local.class": "Local",
	} {
		if got := nestedDeclarationName(name); got != want {
			t.Errorf("nestedDeclarationName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestRevealDocumentSelectsDeclarationAndFocusesEditor(t *testing.T) {
	source := "public class Hello {\n  static class Nested {}\n}"
	doc := &document{source: source, editor: text.NewEditor(source)}
	m := &Model{documents: []*document{doc}, active: 0}
	m.revealDocument(doc, "Nested")
	want := utf8.RuneCountInString(source[:strings.Index(source, "Nested")])
	selection := doc.editor.Selection()
	if doc.editor.Caret() != want || selection.Start != want || selection.End != want+len("Nested") {
		t.Fatalf("declaration selection is %+v with caret %d, want %d..%d", selection, doc.editor.Caret(), want, want+len("Nested"))
	}
	if !m.focusSource {
		t.Fatal("declaration navigation did not request editor focus")
	}
	if !doc.pendingCenter {
		t.Fatal("declaration navigation did not request centered scrolling")
	}
}

func TestRevealDocumentCentersDeclaration(t *testing.T) {
	source := strings.Repeat("// leading context\n", 120) + "public class Hello {}\n" + strings.Repeat("// trailing context\n", 120)
	m := &Model{engine: &fakeEngine{source: source}}
	driver := newDriver(t, m)
	defer driver.Close()
	if err := driver.QueueOpenFiles([]string{writeWorkflowJAR(t)}, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, archiveOpen)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	if err := driver.Advance(20 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	target, ok := m.code.YAtOffset(declarationOffset(source, "Hello"))
	if !ok {
		t.Fatal("declaration has no editor coordinate")
	}
	var viewport float32
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == sourceEditor {
			viewport = hit.Bounds.Height
		}
	}
	got := target - m.code.ScrollY()
	want := viewport / 2
	if got < want-2 || got > want+2 {
		t.Fatalf("declaration is at %.1fpx in the viewport, want center %.1fpx (scroll %.1f)", got, want, m.code.ScrollY())
	}
	if m.currentDocument().pendingCenter {
		t.Fatal("centered declaration scroll was not completed")
	}
}
