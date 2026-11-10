package ui

import (
	"testing"

	"github.com/vivaui/viva/ui"
)

func TestSourceSelectAllAndCopy(t *testing.T) {
	source := "package example;\npublic class Hello {}\n"
	m := &Model{engine: &fakeEngine{source: source}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })

	handled, err := driver.ShortcutChord(ui.ModifierControl, 'a')
	if err != nil {
		t.Fatal(err)
	}
	selection := m.editor.Selection()
	if !handled || selection.Start != 0 || selection.End != m.editor.RuneCount() {
		t.Fatalf("select-all selection = %+v", selection)
	}
	handled, err = driver.ShortcutChord(ui.ModifierControl, 'c')
	if err != nil {
		t.Fatal(err)
	}
	if !handled || driver.Clipboard() != source {
		t.Fatalf("copied text = %q, want the complete source", driver.Clipboard())
	}
}

func TestSourceMouseDragSelectsAndCopies(t *testing.T) {
	source := "package example;\npublic class Hello {}\n"
	m := &Model{engine: &fakeEngine{source: source}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })

	var bounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == sourceEditor {
			bounds = hit.Bounds
		}
	}
	x := bounds.X + 60
	y := bounds.Y + 12
	if err := driver.Press(x, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	for step := 1; step <= 6; step++ {
		if err := driver.Move(x+float32(step)*10, y); err != nil {
			t.Fatal(err)
		}
	}
	if err := driver.Release(x+60, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	selection := m.editor.Selection()
	if m.selectingSource || selection.Start == selection.End {
		t.Fatalf("mouse drag did not select source text: %+v", selection)
	}
	handled, err := driver.ShortcutChord(ui.ModifierControl, 'c')
	if err != nil {
		t.Fatal(err)
	}
	if !handled || driver.Clipboard() == "" {
		t.Fatal("mouse-selected source text was not copied")
	}
}
