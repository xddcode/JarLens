package ui

import (
	"testing"

	"github.com/vivaui/viva/text"
	"github.com/vivaui/viva/ui"
)

func TestFindOccurrences(t *testing.T) {
	matches := findOccurrences("Hello hello HELLO", "hello")
	if len(matches) != 3 {
		t.Fatalf("case-insensitive matches: %+v", matches)
	}
	if matches[0].Start != 0 || matches[1].Start != 6 || matches[2].Start != 12 {
		t.Fatalf("unexpected match offsets: %+v", matches)
	}
	if matches := findOccurrences("aaaa", "aa"); len(matches) != 2 {
		t.Fatalf("matches should not overlap: %+v", matches)
	}
}

func TestGoToLine(t *testing.T) {
	m := &Model{source: "first\nsecond\nthird"}
	m.editor = nil
	m.initialize()
	m.editor = text.NewEditor(m.source)
	if !m.goToLine("2") {
		t.Fatal("goToLine returned false")
	}
	if m.editor.Caret() != len("first\n") {
		t.Fatalf("goToLine did not move to the second line: %d", m.editor.Caret())
	}
	if m.goToLine("invalid") {
		t.Fatal("invalid line was accepted")
	}
}

func TestSourceFindAndGoToLine(t *testing.T) {
	source := "public class Hello {\n  // Hello\n}\n"
	m := &Model{engine: &fakeEngine{source: source}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })

	handled, err := driver.ShortcutChord(ui.ModifierControl, 'f')
	if err != nil || !handled || !m.findOpen {
		t.Fatal("Ctrl+F did not open find")
	}
	for _, control := range []struct {
		id    ui.ID
		label string
	}{{findPrevious, "上一个匹配项"}, {findNext, "下一个匹配项"}, {findClose, "关闭查找"}} {
		if !hasHit(driver, control.id, control.label) {
			t.Errorf("find control %q is not accessible", control.label)
		}
	}
	if err := driver.Type("Hello"); err != nil {
		t.Fatal(err)
	}
	if got := m.findStatus(); got != "1/2" {
		t.Fatalf("initial find status is %q", got)
	}
	if err := driver.Key(ui.KeyEnter, ui.ModifierControl); err != nil {
		t.Fatal(err)
	}
	if got := m.findStatus(); got != "2/2" {
		t.Fatalf("next find status is %q", got)
	}
	if err := driver.Key(ui.KeyEnter, ui.ModifierControl|ui.ModifierShift); err != nil {
		t.Fatal(err)
	}
	if got := m.findStatus(); got != "1/2" {
		t.Fatalf("previous find status is %q", got)
	}
	if err := driver.Key(ui.KeyEscape, 0); err != nil {
		t.Fatal(err)
	}
	if m.findOpen {
		t.Fatal("Escape did not close find")
	}

	handled, err = driver.ShortcutChord(ui.ModifierControl, 'g')
	if err != nil || !handled || !m.goLineOpen {
		t.Fatal("Ctrl+G did not open go to line")
	}
	if err := driver.Type("2"); err != nil {
		t.Fatal(err)
	}
	if err := driver.Key(ui.KeyEnter, 0); err != nil {
		t.Fatal(err)
	}
	if m.goLineOpen || m.editor.Caret() == 0 {
		t.Fatal("go to line did not submit and close")
	}
}
