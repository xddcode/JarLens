package ui

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/decompile"
)

type fakeEngine struct {
	source  string
	sources map[string]string
	err     error
	request chan decompile.Request
	block   bool
	calls   int
}

func (f *fakeEngine) Name() string    { return "Fakeflower" }
func (f *fakeEngine) Version() string { return "1.0" }
func (f *fakeEngine) Decompile(ctx context.Context, request decompile.Request) (decompile.Result, error) {
	f.calls++
	if f.request != nil {
		f.request <- request
	}
	if f.block {
		<-ctx.Done()
		return decompile.Result{}, ctx.Err()
	}
	source := f.source
	if f.sources != nil {
		source = f.sources[request.ClassName]
	}
	return decompile.Result{Source: source, Engine: f.Name(), EngineVersion: f.Version()}, f.err
}

func TestOpenDecompileAndExportWorkflow(t *testing.T) {
	jar := writeWorkflowJAR(t)
	engine := &fakeEngine{source: "package example;\npublic class Hello {}\n"}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	if err := driver.QueueOpenFiles([]string{jar}, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, archiveOpen)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	if !shows(driver, "fixture.jar · 1 classes") {
		t.Fatal("opened archive summary is not visible")
	}

	packageID := findNodeID(t, m, "example")
	click(t, driver, projectNodes.At(packageID))
	if !shows(driver, "Hello") {
		t.Fatal("class did not appear after expanding its package")
	}
	classID := findNodeID(t, m, "Hello.class")
	click(t, driver, projectNodes.At(classID))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	if !shows(driver, "Hello.class") {
		t.Fatal("class tree or tab removed the .class suffix")
	}
	accent := WorkspaceTheme(m.Theme).Accent
	rounded := false
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID != projectNodes.At(classID) {
			continue
		}
		for _, primitive := range driver.Snapshot().Primitives {
			if primitive.Bounds == hit.Bounds && primitive.Color == accent && primitive.Radius == 4 {
				rounded = true
			}
		}
	}
	if !rounded {
		t.Fatal("selected class row is not painted as a rounded accent surface")
	}
	for _, want := range []string{"package example;", "public class Hello {}", "Fakeflower 1.0", "Decompiled by Fakeflower 1.0"} {
		if !shows(driver, want) {
			t.Errorf("workspace does not show %q after decompilation", want)
		}
	}
	click(t, driver, sourceEditor)
	sourceBounds := driver.Snapshot().Hits
	primary := WorkspaceTheme(m.Theme).Primary
	for _, hit := range sourceBounds {
		if hit.ID != sourceEditor {
			continue
		}
		for _, primitive := range driver.Snapshot().Primitives {
			if primitive.Bounds == hit.Bounds && primitive.Color == primary {
				t.Fatal("focused source editor has an accent border")
			}
		}
	}
	if driver.Snapshot().Focused != sourceEditor {
		t.Fatal("source editor did not receive focus")
	}
	if err := driver.Key(ui.KeyDown, 0); err != nil {
		t.Fatal(err)
	}
	if m.editor.Caret() == 0 {
		t.Fatal("source keyboard navigation did not move the caret")
	}

	export := filepath.Join(t.TempDir(), "Hello.java")
	if err := driver.QueueSaveFile(export, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, archiveSave)
	request, ok := driver.LastSaveFileRequest()
	if !ok || request.SuggestedName != "Hello.java" || request.DefaultExtension != ".java" {
		t.Fatalf("class export did not use a Java filename: %+v", request)
	}
	data, err := os.ReadFile(export)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != engine.source {
		t.Fatal("exported source differs from the displayed source")
	}
}

func TestMultipleDocumentsKeepIndependentState(t *testing.T) {
	engine := &fakeEngine{sources: map[string]string{
		"example/Hello.class": "package example;\npublic class Hello {}\n",
		"example/World.class": "package example;\npublic class World {}\n",
	}}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeMultiClassJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	m.editor.MoveTo(1, false)

	click(t, driver, projectNodes.At(findNodeID(t, m, "World.class")))
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/World.class"] })
	if len(m.documents) != 2 || m.active != 1 {
		t.Fatalf("expected two documents with the second active, got %d and %d", len(m.documents), m.active)
	}

	click(t, driver, editorTabs.At(0))
	if m.source != engine.sources["example/Hello.class"] || m.editor.Caret() != 1 {
		t.Fatal("switching tabs did not restore the first document and caret")
	}
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	if engine.calls != 2 {
		t.Fatalf("opening an existing class repeated decompilation: %d calls", engine.calls)
	}

	click(t, driver, editorTabs.At(1))
	click(t, driver, tabClose)
	if len(m.documents) != 1 || m.active != 0 || m.source != engine.sources["example/Hello.class"] {
		t.Fatal("closing the active tab did not reveal the remaining document")
	}
	handled, err := driver.ShortcutChord(ui.ModifierControl, 'w')
	if err != nil {
		t.Fatal(err)
	}
	if !handled || len(m.documents) != 0 || m.active != -1 {
		t.Fatal("Ctrl+W did not close the last document")
	}
}

func TestTextResourceOpensWithoutEngine(t *testing.T) {
	engine := &fakeEngine{}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	content := "<configuration><enabled>true</enabled></configuration>"
	m.OpenPath(writeTextJAR(t, "config/application.xml", content))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "config")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "application.xml")))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	doc := m.currentDocument()
	if doc.language != "xml" || doc.sourceInfo != "Text resource · UTF-8" || engine.calls != 0 {
		t.Fatal("XML resource did not use the text document path")
	}
	if !shows(driver, content) {
		t.Fatal("XML content is not visible in the editor")
	}

	export := filepath.Join(t.TempDir(), "application.xml")
	if err := driver.QueueSaveFile(export, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, archiveSave)
	request, ok := driver.LastSaveFileRequest()
	if !ok || request.SuggestedName != "application.xml" || request.DefaultExtension != ".xml" {
		t.Fatalf("resource export changed its filename: %+v", request)
	}
	data, err := os.ReadFile(export)
	if err != nil || string(data) != content {
		t.Fatalf("resource export differs from source: %q, %v", data, err)
	}
}

func TestNestedClassDisclosureAndOpen(t *testing.T) {
	source := "public class Hello {\n  static class Nested {}\n}\n"
	m := &Model{engine: &fakeEngine{source: source}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeClassJAR(t, "example/Hello.class", "example/Hello$Nested.class"))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	outerID := findNodeID(t, m, "Hello.class")
	if shows(driver, "Nested.class") {
		t.Fatal("nested class appeared before its outer class was expanded")
	}
	click(t, driver, projectDisclosures.At(outerID))
	if !shows(driver, "Nested.class") {
		t.Fatal("outer disclosure did not reveal the nested class")
	}
	nestedID := findNodeID(t, m, "Nested.class")
	click(t, driver, projectNodes.At(nestedID))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	if len(m.documents) != 1 || m.documents[0].entryID != 0 || m.documents[0].name != "Hello.class" {
		t.Fatal("nested class created its own document instead of using the outer class")
	}
	if m.selected != 1 || m.source != source || m.editor.Caret() == 0 {
		t.Fatalf("nested selection did not locate the declaration in the outer source: selected=%d caret=%d source=%q", m.selected, m.editor.Caret(), m.source)
	}
}

func TestProjectSplitterDrag(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	var bounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == projectSplitter {
			bounds = hit.Bounds
			break
		}
	}
	if bounds.Width == 0 {
		t.Fatal("project splitter is not on screen")
	}
	before := m.split.Pos
	x := bounds.X + bounds.Width/2
	y := bounds.Y + 40
	if err := driver.Press(x, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	if err := driver.Move(x+60, y); err != nil {
		t.Fatal(err)
	}
	var current ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == projectSplitter {
			current = hit.Bounds
		}
	}
	theme := WorkspaceTheme(m.Theme)
	var visibleColor ui.Color
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds.Y == current.Y && primitive.Bounds.Height == current.Height &&
			primitive.Bounds.X >= current.X && primitive.Bounds.X < current.X+current.Width {
			visibleColor = primitive.Color
		}
	}
	if visibleColor != theme.Border {
		t.Fatalf("dragged splitter has a non-neutral indicator: %+v", visibleColor)
	}
	if err := driver.Release(x+60, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	if m.split.Pos <= before {
		t.Fatalf("splitter did not move: before %v, after %v", before, m.split.Pos)
	}
}

func TestOpenArchiveByDrop(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	accepted, err := driver.Drop(ui.FileDropEvent{Phase: ui.FileDropDrop, Paths: []string{writeWorkflowJAR(t)}})
	if err != nil {
		t.Fatal(err)
	}
	if !accepted {
		t.Fatal("workspace refused a JAR drop")
	}
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
}

func TestOpenFailureKeepsWorkspaceUsable(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	if err := driver.QueueOpenFiles([]string{"invalid.txt"}, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, archiveOpen)
	pumpUntil(t, driver, func() bool { return m.diagnostic != "" })
	if !shows(driver, "Only JAR and Class files are supported") {
		t.Fatal("unsupported input did not produce a useful message")
	}
}

func TestCancelIgnoresLateEngineResult(t *testing.T) {
	request := make(chan decompile.Request, 1)
	engine := &fakeEngine{source: "too late", request: request, block: true}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	m.startOpen(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	select {
	case <-request:
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not receive the request")
	}
	click(t, driver, workCancel)
	if !shows(driver, "Operation cancelled") {
		t.Fatal("cancellation is not visible")
	}
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	if m.source != "" || m.busy {
		t.Fatal("cancelled task changed the source or remained busy")
	}
}

func TestOpenPathOpensArchive(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
}

func pumpUntil(t *testing.T, driver *app.Driver, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := driver.Pump(); err != nil {
			t.Fatal(err)
		}
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("background operation did not reach the expected state")
}

func findNodeID(t *testing.T, m *Model, name string) int {
	t.Helper()
	for _, node := range m.workspace.Tree.Nodes {
		if node.Name == name {
			return node.ID
		}
	}
	t.Fatalf("node %q not found", name)
	return 0
}

func writeWorkflowJAR(t *testing.T) string {
	t.Helper()
	return writeClassJAR(t, "example/Hello.class")
}

func writeMultiClassJAR(t *testing.T) string {
	t.Helper()
	return writeClassJAR(t, "example/Hello.class", "example/World.class")
}

func writeClassJAR(t *testing.T, classes ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, name := range classes {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 65}); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTextJAR(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
