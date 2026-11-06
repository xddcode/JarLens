package ui

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type blockingEngine struct {
	requests chan decompile.Request
}

func (e *blockingEngine) Name() string    { return "Blocking" }
func (e *blockingEngine) Version() string { return "1.0" }
func (e *blockingEngine) Decompile(ctx context.Context, request decompile.Request) (decompile.Result, error) {
	e.requests <- request
	<-ctx.Done()
	return decompile.Result{}, ctx.Err()
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
	if !shows(driver, "fixture.jar · 1 个类") {
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
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	if !shows(driver, "Hello.class") {
		t.Fatal("class tree or tab removed the .class suffix")
	}
	selection := WorkspaceTheme(m.Theme).Accent
	rounded := false
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID != projectNodes.At(classID) {
			continue
		}
		for _, primitive := range driver.Snapshot().Primitives {
			if primitive.Bounds == hit.Bounds && primitive.Color == selection && primitive.Radius == 5 {
				rounded = true
			}
		}
	}
	if !rounded {
		t.Fatal("selected class row does not use the Islands selection surface")
	}
	for _, want := range []string{"package example;", "public class Hello {}", "Fakeflower 1.0", "反编译引擎：Fakeflower 1.0"} {
		if !shows(driver, want) {
			t.Errorf("workspace does not show %q after decompilation", want)
		}
	}
	classSelection := m.editor.Selection()
	if got := m.source[m.editor.ByteOffset(classSelection.Start):m.editor.ByteOffset(classSelection.End)]; got != "Hello" {
		t.Fatalf("outer class name was not selected after opening: %q", got)
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

func TestProjectTitleMenuControlsTree(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeMultiClassJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })

	click(t, driver, projectTitle)
	click(t, driver, projectExpandAll)
	if !m.expanded[findNodeID(t, m, "example")] {
		t.Fatal("project title menu did not expand the tree")
	}
	click(t, driver, projectTitle)
	click(t, driver, projectCollapseAll)
	if len(m.expanded) != 0 {
		t.Fatal("project title menu did not collapse the tree")
	}
	click(t, driver, projectTitle)
	click(t, driver, projectFilterMenu)
	if !m.filterOpen || m.projectMenuOpen {
		t.Fatal("project title menu did not open the filter and close")
	}
}

func TestProjectFilter(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeMultiClassJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectSearch)
	click(t, driver, projectFilter)
	if err := driver.Type("Hello"); err != nil {
		t.Fatal(err)
	}
	if !shows(driver, "Hello.class") || shows(driver, "World.class") {
		t.Fatal("project filter did not keep only the matching class")
	}
	m.filter.SetText("missing")
	if err := driver.Move(0, 0); err != nil {
		t.Fatal(err)
	}
	if !shows(driver, "没有匹配的类或资源") {
		t.Fatal("empty filter result was not explained")
	}
	m.filter.SetText("")
	if err := driver.Key(ui.KeyTab, 0); err != nil {
		t.Fatal(err)
	}
	if !shows(driver, "World.class") {
		t.Fatal("clearing the filter did not restore the tree")
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
	selectedBackground, selectedBorder, _ := editorTabColors(m.Theme)
	selectedPill, border := false, false
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == editorTabClose.At(1) && hit.Hint != "关闭 World.class (Ctrl+W)" {
			t.Errorf("close tooltip does not describe the available shortcut: %q", hit.Hint)
		}
		if hit.ID != editorTabs.At(1) {
			continue
		}
		if want := m.workspace.Archive.Path + "!/example/World.class"; hit.Hint != want {
			t.Errorf("tab tooltip is %q instead of the document location %q", hit.Hint, want)
		}
		for _, primitive := range driver.Snapshot().Primitives {
			inner := hit.Bounds
			inner.X++
			inner.Y++
			inner.Width -= 2
			inner.Height -= 2
			if primitive.Bounds == inner && primitive.Radius == 5 && primitive.Color == selectedBackground {
				selectedPill = true
			}
			if primitive.Bounds == hit.Bounds && primitive.Radius == 6 && primitive.Color == selectedBorder {
				border = true
			}
			if primitive.Bounds.Height == 2 && primitive.Color == WorkspaceTheme(m.Theme).Primary {
				t.Fatal("editor tab uses an underline instead of the OpenIDEA selected pill")
			}
		}
	}
	if !selectedPill || !border {
		t.Fatal("active editor tab does not have its rounded selection and inset border")
	}
	snapshot := driver.Snapshot()
	var tabBounds, closeBounds ui.Rect
	for _, hit := range snapshot.Hits {
		switch hit.ID {
		case editorTabs.At(1):
			tabBounds = hit.Bounds
		case editorTabClose.At(1):
			closeBounds = hit.Bounds
		}
	}
	labelFound := false
	for _, label := range snapshot.Labels {
		if label.Value != "World.class" || label.Bounds.X < tabBounds.X || label.Bounds.X >= tabBounds.X+tabBounds.Width {
			continue
		}
		labelFound = true
		delta := label.Bounds.Y + label.Bounds.Height/2 - (tabBounds.Y + tabBounds.Height/2)
		if delta < -0.5 || delta > 0.5 {
			t.Errorf("editor tab label is not vertically centered: label %+v, tab %+v", label.Bounds, tabBounds)
		}
	}
	if !labelFound {
		t.Fatal("active editor tab label was not painted")
	}
	if delta := closeBounds.Y + closeBounds.Height/2 - (tabBounds.Y + tabBounds.Height/2); delta < -0.5 || delta > 0.5 {
		t.Errorf("editor tab close button is not vertically centered: close %+v, tab %+v", closeBounds, tabBounds)
	}

	click(t, driver, editorTabs.At(0))
	if m.source != engine.sources["example/Hello.class"] || m.editor.Caret() != 1 {
		t.Fatal("switching tabs did not restore the first document and caret")
	}
	if m.status != "已打开 Hello.class" {
		t.Fatalf("switching tabs left a stale status message: %q", m.status)
	}
	if err := driver.Move(closeBounds.X+closeBounds.Width/2, closeBounds.Y+closeBounds.Height/2); err != nil {
		t.Fatal(err)
	}
	_, _, tabHover := editorTabColors(m.Theme)
	inactiveHoverPainted := false
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds == tabBounds && primitive.Radius == 6 && primitive.Color == tabHover {
			inactiveHoverPainted = true
		}
	}
	if !inactiveHoverPainted {
		t.Fatal("hovering an inactive tab's close action removes the tab hover background")
	}
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	if engine.calls != 2 {
		t.Fatalf("opening an existing class repeated decompilation: %d calls", engine.calls)
	}

	click(t, driver, editorTabs.At(1))
	if err := driver.Move(closeBounds.X+closeBounds.Width/2, closeBounds.Y+closeBounds.Height/2); err != nil {
		t.Fatal(err)
	}
	hoverPainted := false
	for _, primitive := range driver.Snapshot().Primitives {
		bounds := primitive.Bounds
		if bounds.Width == 16 && bounds.Height == 16 && primitive.Radius == 8 && primitive.Color == WorkspaceTheme(m.Theme).Primary &&
			bounds.X+bounds.Width/2 == closeBounds.X+closeBounds.Width/2 && bounds.Y+bounds.Height/2 == closeBounds.Y+closeBounds.Height/2 {
			hoverPainted = true
		}
	}
	if !hoverPainted {
		t.Fatal("hovered tab close action does not show its circular accent feedback")
	}
	click(t, driver, editorTabClose.At(1))
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

func TestTabContextMenuSplitsEditors(t *testing.T) {
	m := &Model{engine: &fakeEngine{source: "package example;\npublic class Hello {\n  void greet() {}\n}\n"}}
	driver := newDriver(t, m)
	defer driver.Close()

	openHello := func() {
		previous := m.workspace
		m.OpenPath(writeWorkflowJAR(t))
		pumpUntil(t, driver, func() bool { return m.workspace != previous })
		click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
		click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
		pumpUntil(t, driver, func() bool { return m.source != "" })
	}
	rightClickTab := func() {
		for _, hit := range driver.Snapshot().Hits {
			if hit.ID == editorTabs.At(0) {
				if err := driver.ClickButton(hit.Bounds.X+hit.Bounds.Width/2, hit.Bounds.Y+hit.Bounds.Height/2, ui.PointerButtonSecondary); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		t.Fatal("editor tab is not painted")
	}
	openHello()
	rightClickTab()
	for _, label := range []string{"关闭", "关闭其他标签页", "关闭全部标签页", "关闭左侧标签页", "关闭右侧标签页", "向右拆分", "向下拆分"} {
		if !shows(driver, label) {
			t.Errorf("tab context menu does not show %q", label)
		}
	}
	if err := driver.Key(ui.KeyEscape, 0); err != nil {
		t.Fatal(err)
	}
	if m.tabsContextOpen {
		t.Fatal("Escape did not close the tab context menu")
	}
	rightClickTab()
	click(t, driver, editorTabSplitRight)
	if m.secondary == nil || m.splitAxis != splitRight {
		t.Fatal("Split Right did not create a secondary editor")
	}
	var primaryBounds, secondaryBounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		switch hit.ID {
		case sourceEditor:
			primaryBounds = hit.Bounds
		case splitSourceEditor:
			secondaryBounds = hit.Bounds
		}
	}
	if secondaryBounds.X <= primaryBounds.X+primaryBounds.Width {
		t.Fatalf("right split is not beside the primary editor: %+v %+v", primaryBounds, secondaryBounds)
	}
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	if focused := driver.Snapshot().Focused; focused != splitSourceEditor {
		t.Fatalf("new split editor did not receive focus: %v", focused)
	}
	primaryCaret := m.editor.Caret()
	click(t, driver, splitSourceEditor)
	before := m.secondary.editor.Caret()
	if err := driver.Key(ui.KeyDown, 0); err != nil {
		t.Fatal(err)
	}
	if m.secondary.editor.Caret() == before || m.editor.Caret() != primaryCaret {
		t.Fatal("split editor caret is not independent of the primary editor")
	}

	m.closeDocument(0)
	if err := driver.Move(0, 0); err != nil {
		t.Fatal(err)
	}
	if m.secondary != nil {
		t.Fatal("closing the document retained its split model")
	}
	if hasHit(driver, splitSourceEditor, "Hello.class") {
		t.Fatal("closing the document retained its split editor hit")
	}
	openHello()
	rightClickTab()
	click(t, driver, editorTabSplitDown)
	if m.secondary == nil || m.splitAxis != splitDown {
		t.Fatal("Split Down did not create a secondary editor")
	}
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == sourceEditor {
			primaryBounds = hit.Bounds
		}
		if hit.ID == splitSourceEditor {
			secondaryBounds = hit.Bounds
		}
	}
	if secondaryBounds.Y <= primaryBounds.Y+primaryBounds.Height {
		t.Fatalf("down split is not below the primary editor: %+v %+v", primaryBounds, secondaryBounds)
	}
	rightClickTab()
	click(t, driver, editorTabCloseCurrent)
	if len(m.documents) != 0 || m.secondary != nil {
		t.Fatal("context-menu Close did not close the document and its split")
	}
}

func TestEditorTabMenuCloseActions(t *testing.T) {
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
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/Hello.class"] })
	click(t, driver, projectNodes.At(findNodeID(t, m, "World.class")))
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/World.class"] })

	click(t, driver, editorTabsMenuOpen)
	for _, label := range []string{"关闭全部标签页", "关闭其他标签页", "关闭左侧标签页", "关闭右侧标签页"} {
		if !shows(driver, label) {
			t.Errorf("tab menu does not show %q", label)
		}
	}
	var firstMenuRow, moreBounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		switch hit.ID {
		case editorTabsCloseAll:
			firstMenuRow = hit.Bounds
			if hit.Bounds.Height != 24 {
				t.Errorf("tab menu row height is %v instead of 24", hit.Bounds.Height)
			}
		case editorTabsMenuOpen:
			moreBounds = hit.Bounds
		}
	}
	if focused := driver.Snapshot().Focused; focused != editorTabsCloseAll {
		t.Fatalf("tab menu did not focus its first action: %v", focused)
	}
	pressedPainted := false
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds == moreBounds && primitive.Color == toolbarPressedColor(m.Theme) && primitive.Radius == 6 {
			pressedPainted = true
		}
	}
	if !pressedPainted {
		t.Fatal("open tab-actions button does not retain its pressed appearance")
	}
	initialRowPainted := false
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds == firstMenuRow && primitive.Color == WorkspaceTheme(m.Theme).Accent && primitive.Radius == 5 {
			initialRowPainted = true
		}
	}
	if !initialRowPainted {
		t.Fatal("opening the tab menu does not show the first action as current")
	}
	if err := driver.Move(firstMenuRow.X+firstMenuRow.Width/2, firstMenuRow.Y+firstMenuRow.Height/2); err != nil {
		t.Fatal(err)
	}
	currentRowPainted := false
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds == firstMenuRow && primitive.Color == WorkspaceTheme(m.Theme).Accent && primitive.Radius == 5 {
			currentRowPainted = true
		}
	}
	if !currentRowPainted {
		t.Fatal("current tab-menu row does not use the Islands selection pill")
	}
	if err := driver.Key(ui.KeyEscape, 0); err != nil {
		t.Fatal(err)
	}
	if m.tabsMenuOpen {
		t.Fatal("Escape did not close the tab menu")
	}

	click(t, driver, editorTabsMenuOpen)
	if err := driver.Click(2, 200); err != nil {
		t.Fatal(err)
	}
	if m.tabsMenuOpen {
		t.Fatal("outside click did not close the tab menu")
	}

	click(t, driver, editorTabsMenuOpen)
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == editorTabsCloseRight && !hit.Disabled {
			t.Error("Close Tabs to the Right is enabled for the last tab")
		}
		if hit.ID == editorTabsCloseLeft && hit.Disabled {
			t.Error("Close Tabs to the Left is disabled with a tab to the left")
		}
	}
	click(t, driver, editorTabsCloseOthers)
	if len(m.documents) != 1 || m.active != 0 || m.currentDocument().name != "World.class" || m.tabsMenuOpen {
		t.Fatal("Close Other Tabs did not keep only the active document")
	}

	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/Hello.class"] })
	click(t, driver, editorTabsMenuOpen)
	click(t, driver, editorTabsCloseLeft)
	if len(m.documents) != 1 || m.currentDocument().name != "Hello.class" || m.active != 0 {
		t.Fatal("Close Tabs to the Left removed the active document")
	}

	click(t, driver, projectNodes.At(findNodeID(t, m, "World.class")))
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/World.class"] })
	click(t, driver, editorTabs.At(0))
	click(t, driver, editorTabsMenuOpen)
	click(t, driver, editorTabsCloseRight)
	if len(m.documents) != 1 || m.currentDocument().name != "Hello.class" || m.source != engine.sources["example/Hello.class"] {
		t.Fatal("Close Tabs to the Right did not preserve the active source")
	}

	click(t, driver, projectNodes.At(findNodeID(t, m, "World.class")))
	pumpUntil(t, driver, func() bool { return m.source == engine.sources["example/World.class"] })
	click(t, driver, editorTabsMenuOpen)
	click(t, driver, editorTabsCloseAll)
	if len(m.documents) != 0 || m.active != -1 || m.source != "" || m.tabsMenuOpen {
		t.Fatal("Close All Tabs did not clear the editor")
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
	if doc.language != "xml" || doc.sourceInfo != "文本资源 · UTF-8" || engine.calls != 0 {
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

func TestClassLinksNavigateToServiceAndJavaReferences(t *testing.T) {
	engine := &fakeEngine{sources: map[string]string{
		"example/Hello.class": "public class Hello {}\n",
		"example/World.class": "public class World { Hello hello; }\n",
	}}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeLinkedJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "META-INF")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "services")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "example.Greeter")))
	pumpUntil(t, driver, func() bool { return m.source == "example.World\n" })
	underlined := false
	for _, text := range driver.Snapshot().Texts {
		if text.Value != "example.World" {
			continue
		}
		for _, run := range driver.Snapshot().TextRuns(text) {
			underlined = underlined || run.Underline
		}
	}
	if !underlined {
		t.Fatal("class link is not underlined")
	}
	for _, label := range driver.Snapshot().Labels {
		if label.Value != "example.World" {
			continue
		}
		if err := driver.Move(label.Bounds.X+label.Bounds.Width/2, label.Bounds.Y+label.Bounds.Height/2); err != nil {
			t.Fatal(err)
		}
	}
	if cursor := driver.Snapshot().Cursor; cursor != app.CursorHand {
		t.Fatalf("hovered class link cursor is %v instead of a hand", cursor)
	}
	clickSourceText(t, driver, "example.World", "example.World")
	pumpUntil(t, driver, func() bool {
		return m.currentDocument() != nil && m.currentDocument().name == "World.class" && m.source == engine.sources["example/World.class"]
	})
	selection := m.editor.Selection()
	if got := m.source[m.editor.ByteOffset(selection.Start):m.editor.ByteOffset(selection.End)]; got != "World" {
		t.Fatalf("service link did not select the target declaration: %q", got)
	}
	worldID := findNodeID(t, m, "World.class")
	if parent := m.workspace.Tree.Nodes[worldID].Parent; !m.expanded[parent] {
		t.Fatal("target class parent was not expanded after navigation")
	}
	selectedInTree := false
	var selectedBounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == projectNodes.At(worldID) && hit.Selected {
			selectedInTree = true
			selectedBounds = hit.Bounds
		}
	}
	if !selectedInTree {
		t.Fatal("target class was not selected in the project tree")
	}
	blueSelection := false
	for _, primitive := range driver.Snapshot().Primitives {
		blueSelection = blueSelection || primitive.Bounds == selectedBounds && primitive.Color == WorkspaceTheme(m.Theme).Accent && primitive.Radius == 5
	}
	if !blueSelection {
		t.Fatal("navigated class does not use the blue selected background")
	}
	clickSourceText(t, driver, "public class World { Hello hello; }", "Hello")
	pumpUntil(t, driver, func() bool {
		return m.currentDocument() != nil && m.currentDocument().name == "Hello.class" && m.source == engine.sources["example/Hello.class"]
	})
}

func clickSourceText(t *testing.T, driver *app.Driver, line, target string) {
	t.Helper()
	for _, label := range driver.Snapshot().Labels {
		if label.Value != line {
			continue
		}
		at := strings.Index(line, target)
		if at < 0 {
			break
		}
		x := label.Bounds.X + label.Bounds.Width*float32(at+len(target)/2)/float32(len(line))
		if err := driver.Click(x, label.Bounds.Y+label.Bounds.Height/2); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatalf("source line %q is not painted", line)
}

func TestProjectTreePaintsOnlyVisibleRows(t *testing.T) {
	classes := make([]string, 240)
	for index := range classes {
		classes[index] = "example/Class" + strconv.Itoa(index) + ".class"
	}
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeClassJAR(t, classes...))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	painted := 0
	for _, label := range driver.Snapshot().Labels {
		if strings.HasPrefix(label.Value, "Class") && strings.HasSuffix(label.Value, ".class") {
			painted++
		}
	}
	if painted == 0 || painted >= len(classes) {
		t.Fatalf("project tree painted %d of %d rows instead of only the viewport", painted, len(classes))
	}
	if err := driver.Scroll(ui.ScrollEvent{X: 100, Y: 300, DeltaY: 10000}); err != nil {
		t.Fatal(err)
	}
	visible := m.workspace.Tree.Visible(m.expanded)
	lastName := visible[len(visible)-1].Node.Name
	if !shows(driver, lastName) {
		t.Fatalf("virtualized tree did not paint its last row %q after scrolling", lastName)
	}
}

func TestOpeningDocumentScrollsProjectTreeIntoView(t *testing.T) {
	classes := make([]string, 40)
	for index := range classes {
		classes[index] = "example/Class" + strconv.Itoa(index) + ".class"
	}
	m := &Model{engine: &fakeEngine{source: "public class Example {}\n"}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeClassJAR(t, classes...))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	last := len(m.workspace.Archive.Entries) - 1
	m.startDecompile(last)
	pumpUntil(t, driver, func() bool { return m.source != "" })
	if !m.expanded[m.workspace.Tree.Nodes[findNodeID(t, m, "Class39.class")].Parent] {
		t.Fatal("opening a document did not expand its package")
	}
	if offset := m.view.ScrollState(projectTree).Offset; offset <= 0 {
		t.Fatalf("opening an offscreen document did not scroll the tree: %v", offset)
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
	selection := m.editor.Selection()
	if got := m.source[m.editor.ByteOffset(selection.Start):m.editor.ByteOffset(selection.End)]; got != "Nested" {
		t.Fatalf("nested declaration name is not selected: %q", got)
	}
	if focused := driver.Snapshot().Focused; focused != sourceEditor {
		t.Fatalf("nested navigation did not focus the source editor: %v", focused)
	}
}

func TestProjectTreeLongNamesScrollHorizontally(t *testing.T) {
	const name = "ArchetypeEmbodimentIntegrationConfigWithLongNestedName.class"
	m := &Model{engine: &fakeEngine{source: "public class Example {}\n"}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeClassJAR(t, "org/example/"+name))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "org")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	if !shows(driver, name) {
		t.Fatal("long class name was truncated instead of keeping its natural width")
	}
	var row ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == projectNodes.At(findNodeID(t, m, name)) {
			row = hit.Bounds
		}
	}
	labelWidth := float32(0)
	for _, label := range driver.Snapshot().Labels {
		if label.Value == name {
			labelWidth = label.Bounds.Width
		}
	}
	if labelWidth < 200 {
		t.Fatalf("long class label was squeezed: width %v", labelWidth)
	}
	if !hasHit(driver, projectHorizontalScrollbar, "横向滚动项目树") {
		t.Fatal("wide project tree did not provide a horizontal scrollbar")
	}
	if err := driver.Scroll(ui.ScrollEvent{X: row.X + 20, Y: row.Y + 12, DeltaX: 60, Precise: true}); err != nil {
		t.Fatal(err)
	}
	if m.treeScrollX <= 0 {
		t.Fatal("horizontal wheel did not move the project tree")
	}
	if err := driver.Pump(); err != nil {
		t.Fatal(err)
	}
	geometry := m.treeScrollGeometry
	if geometry.HorizontalThumb.Width == 0 {
		t.Fatal("horizontal scrollbar geometry was not published")
	}
	x := geometry.HorizontalThumb.X + geometry.HorizontalThumb.Width/2
	y := geometry.HorizontalThumb.Y + geometry.HorizontalThumb.Height/2
	before := m.treeScrollX
	if err := driver.Press(x, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	if err := driver.Move(x+20, y); err != nil {
		t.Fatal(err)
	}
	if m.treeScrollX == before {
		t.Fatal("dragging the horizontal scrollbar did not move the tree")
	}
	if err := driver.Release(x+20, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	click(t, driver, projectNodes.At(findNodeID(t, m, name)))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	if !shows(driver, "CL") {
		t.Fatal("class badge text was clipped")
	}
	wideSelection := false
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds.Width > m.split.Pos && primitive.Radius == 5 {
			wideSelection = true
		}
	}
	if !wideSelection {
		t.Fatal("selected tree row does not cover its full natural width")
	}
}

func TestProjectSplitterDrag(t *testing.T) {
	m := &Model{engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })

	var bounds ui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == projectSplitter {
			bounds = hit.Bounds
			break
		}
	}
	if bounds.Width != 6 {
		t.Fatalf("project splitter gap is %v pixels instead of the Islands 6-pixel gap", bounds.Width)
	}
	before := m.split.Pos
	x := bounds.X + bounds.Width/2
	y := bounds.Y + 40
	if err := driver.Press(x, y, ui.PointerButtonPrimary); err != nil {
		t.Fatal(err)
	}
	for step := 1; step <= 12; step++ {
		if err := driver.Move(x+float32(step)*5, y); err != nil {
			t.Fatal(err)
		}
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
	maximizes := 0
	m := &Model{engine: &fakeEngine{}, MaximizeWindow: func() { maximizes++ }}
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
	if maximizes != 1 {
		t.Fatalf("JAR drop maximized the window %d times, want 1", maximizes)
	}
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
	if !shows(driver, "仅支持 JAR 和 Class 文件") {
		t.Fatal("unsupported input did not produce a useful message")
	}
}

func TestRapidDocumentSwitchCanRestartCancelledClass(t *testing.T) {
	engine := &blockingEngine{requests: make(chan decompile.Request, 4)}
	m := &Model{engine: engine}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeMultiClassJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	waitEngineRequest(t, engine.requests)
	click(t, driver, projectNodes.At(findNodeID(t, m, "World.class")))
	waitEngineRequest(t, engine.requests)
	if m.documents[0].loading {
		t.Fatal("cancelled Hello document remained in its loading state")
	}
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	waitEngineRequest(t, engine.requests)
	if !m.documents[0].loading || m.currentDocument().name != "Hello.class" {
		t.Fatal("returning to the cancelled class did not restart decompilation")
	}
}

func waitEngineRequest(t *testing.T, requests <-chan decompile.Request) {
	t.Helper()
	select {
	case <-requests:
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not receive the document request")
	}
}

func TestLatestTaskResultIsNotOverwritten(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()

	m.publish(&taskResult{generation: 2})
	m.publish(&taskResult{generation: 1})
	m.mu.Lock()
	generation := m.pending.generation
	m.mu.Unlock()
	if generation != 2 {
		t.Fatalf("older task replaced the pending latest result: generation %d", generation)
	}
}

func TestOpenArchiveDismissesTabMenuWithoutPanic(t *testing.T) {
	m := &Model{engine: &fakeEngine{source: "public class Hello {}\n"}}
	driver := newDriver(t, m)
	defer driver.Close()

	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	click(t, driver, projectNodes.At(findNodeID(t, m, "example")))
	click(t, driver, projectNodes.At(findNodeID(t, m, "Hello.class")))
	pumpUntil(t, driver, func() bool { return m.source != "" })
	click(t, driver, editorTabsMenuOpen)
	previous := m.workspace
	m.OpenPath(writeMultiClassJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != previous })
	if m.tabsMenuOpen || len(m.documents) != 0 {
		t.Fatal("opening another archive retained stale tab-menu or document state")
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
	if !shows(driver, "操作已取消") {
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

func writeLinkedJAR(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, name := range []string{"example/Hello.class", "example/World.class"} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 65}); err != nil {
			t.Fatal(err)
		}
	}
	entry, err := writer.Create("META-INF/services/example.Greeter")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("example.World\n")); err != nil {
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
