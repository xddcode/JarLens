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
	err     error
	request chan decompile.Request
	block   bool
}

func (f *fakeEngine) Name() string    { return "Fakeflower" }
func (f *fakeEngine) Version() string { return "1.0" }
func (f *fakeEngine) Decompile(ctx context.Context, request decompile.Request) (decompile.Result, error) {
	if f.request != nil {
		f.request <- request
	}
	if f.block {
		<-ctx.Done()
		return decompile.Result{}, ctx.Err()
	}
	return decompile.Result{Source: f.source, Engine: f.Name(), EngineVersion: f.Version()}, f.err
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
	classID := findNodeID(t, m, "Hello")
	click(t, driver, projectNodes.At(classID))
	pumpUntil(t, driver, func() bool { return m.source != "" })
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
	data, err := os.ReadFile(export)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != engine.source {
		t.Fatal("exported source differs from the displayed source")
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
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello")))
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
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("example/Hello.class")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 65}); err != nil {
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
