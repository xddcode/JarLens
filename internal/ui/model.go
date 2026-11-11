package ui

import (
	"context"
	"errors"

	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/text"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
	"github.com/xddcode/jarlens/internal/decompile"
	jarlensupdate "github.com/xddcode/jarlens/internal/update"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

const repositoryURL = "https://github.com/xddcode/jarlens"

type taskKind uint8

const (
	taskOpen taskKind = iota
	taskDecompile
	taskResource
	taskUpdate
)

type taskResult struct {
	kind       taskKind
	generation uint64
	entryID    int
	pane       *editorPane
	workspace  *jarlensworkspace.Workspace
	result     decompile.Result
	update     *jarlensupdate.Release
	automatic  bool
	err        error
}

type document struct {
	entryID             int
	name                string
	path                string
	location            string
	language            string
	source              string
	sourceInfo          string
	diagnostic          string
	revealName          string
	pendingCenter       bool
	pendingCenterOffset int
	loading             bool
	links               []classLink
	editor              *text.Editor
	code                components.CodeEditor
}

type treeLayoutCache struct {
	tree     *archive.Tree
	query    string
	expanded uint64
	width    float32
}

type splitDirection uint8

const (
	splitRight splitDirection = iota
	splitDown
)

type editorPane struct {
	id        int
	parent    *editorPane
	documents []*document
	active    int
	editor    *text.Editor
	code      components.CodeEditor
	axis      splitDirection
	splitter  flex.SplitterState
	width     float32
	height    float32
	first     *editorPane
	second    *editorPane
}

type RecentStore interface {
	Load() []string
	Record(string) []string
}

type SettingsStore interface {
	Load() jarlensworkspace.Settings
	Save(jarlensworkspace.Settings)
}

type Model struct {
	Theme                Appearance
	TrayIcon             app.WindowIcon
	RecentStore          RecentStore
	SettingsStore        SettingsStore
	MaximizeWindow       func()
	RestoreWindow        func()
	UpdateChecker        func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error)
	UpdateReleasesLoader func(context.Context, jarlensupdate.Source, int, int) (jarlensupdate.ReleasePage, error)
	SetLaunchAtLogin     func(bool) error

	updateRelease            *jarlensupdate.Release
	updateChecking           bool
	updateDialogOpen         bool
	updateStatus             string
	updateCancel             context.CancelFunc
	updateReleases           []jarlensupdate.Release
	updateReleasesLoading    bool
	updateReleasesLoaded     bool
	updateReleasesHasMore    bool
	updateReleasesPage       int
	updateReleasesSource     jarlensupdate.Source
	updateReleasesError      string
	updateReleasesGeneration uint64
	updateReleasesCancel     context.CancelFunc
	updateReleasesRequested  bool
	updateReleaseExpanded    map[string]bool
	pendingUpdateReleases    *updateReleasesResult

	settings                   jarlensworkspace.Settings
	settingsDraft              jarlensworkspace.Settings
	settingsJava               flex.Field
	settingsFontSize           flex.Field
	settingsSearch             flex.Field
	settingsThemeSelect        flex.SelectState
	settingsLanguageSelect     flex.SelectState
	settingsUpdateSourceSelect flex.SelectState
	settingsEngineSelect       flex.SelectState
	settingsOpen               bool
	settingsCategory           int
	settingsQuery              string
	settingsJavaText           string
	settingsFontSizeText       string
	settingsError              string
	settingsAppearanceExpanded bool
	settingsEditorExpanded     bool
	settingsHistory            []int
	settingsHistoryIndex       int
	fontPickerOpen             bool
	fontPickerTarget           int
	fontPickerSearch           flex.Field
	fontPickerFamilies         []string

	view            flex.View
	filter          flex.Field
	findField       flex.Field
	goLineField     flex.Field
	tooltip         workspaceTooltip
	code            components.CodeEditor
	codeFace        vivaui.MonoFace
	appIcon         vivaui.Bitmap
	split           flex.SplitterState
	editor          *text.Editor
	engine          decompile.Engine
	engines         []decompile.Engine
	workspace       *jarlensworkspace.Workspace
	expanded        map[int]bool
	documents       []*document
	active          int
	selected        int
	source          string
	sourceInfo      string
	busy            bool
	status          string
	diagnostic      string
	filterOpen      bool
	filterFocus     bool
	projectMenuOpen bool
	findOpen        bool
	findFocus       bool
	findQuery       string
	findSource      string
	findMatches     []text.Span
	findIndex       int
	goLineOpen      bool
	tabsMenuOpen    bool
	tabsMenuPane    *editorPane
	focusSource     bool
	focusProject    bool
	focusPane       *editorPane

	paneRoot          *editorPane
	activePane        *editorPane
	nextPaneID        int
	tabsContextOpen   bool
	tabsContextPane   *editorPane
	tabsContextTab    int
	tabsContextPoint  vivaui.Point
	mainMenuOpen      bool
	mainMenuFromTitle bool
	mainMenuTitleRect vivaui.Rect
	homeVisible       bool
	recentProjects    []string
	projectMenuScroll components.MenuScroll

	treeLayout          treeLayoutCache
	treeScrollX         float32
	treeScrollGeometry  components.ScrollAreaGeometry
	treeScrollDragging  bool
	treeScrollStartX    float32
	treeScrollStartPos  float32
	classLinks          *classLinkIndex
	linkDownOffset      int
	sourcePointer       vivaui.Point
	sourcePointerInside bool
	selectingSource     bool

	invalidator           *app.Invalidator
	application           *app.Application
	window                *app.Window
	statusItem            *app.StatusItem
	EnsureTray            func() bool
	HideWindow            func()
	stopSystemAppearance  func()
	nativeAppearance      app.Appearance
	nativeAppearanceDirty bool
	mu                    sync.Mutex
	pending               *taskResult
	generation            uint64
	taskEntry             int
	taskPane              *editorPane
	cancel                context.CancelFunc
	openURL               func(string) error
}

func (m *Model) initialize() {
	configureFonts()
	m.codeFace = monoFace
	m.appIcon = applicationIcon()
	if m.RecentStore != nil {
		m.recentProjects = m.RecentStore.Load()
	}
	m.settings = jarlensworkspace.DefaultSettings()
	if m.SettingsStore != nil {
		m.settings = m.SettingsStore.Load()
	}
	if m.split.Pos == 0 {
		m.split.Pos = 300
	}
	if m.editor == nil {
		m.editor = text.NewEditor("")
	}
	m.initializeEngines()
	m.applySettings(false)
	if m.expanded == nil {
		m.expanded = make(map[int]bool)
	}
	if len(m.documents) == 0 {
		m.active = -1
	}
	if m.selected == 0 && m.workspace == nil {
		m.selected = -1
	}
	m.taskEntry = -1
	m.linkDownOffset = -1
	m.tabsContextTab = -1
	m.ensurePane()
}

func (m *Model) OpenPath(path string) {
	m.openPath(path, true)
}

func (m *Model) RestorePath(path string, maximize bool) {
	m.openPath(path, maximize)
}

func (m *Model) chooseArchive() {
	paths, err := app.OpenFiles(app.OpenFileOptions{
		Title:   m.t("打开 Java 归档"),
		Filters: []app.FileFilter{{Name: m.t("Java 归档和类文件"), Pattern: "*.jar;*.class"}},
	})
	if err != nil {
		m.fail("无法打开文件选择对话框", err)
		return
	}
	if len(paths) > 0 {
		m.startOpen(paths[0])
	}
}

func (m *Model) startOpen(path string) {
	m.openPath(path, true)
}

func (m *Model) openPath(path string, maximize bool) {
	if maximize && (m.workspace == nil || m.homeVisible) && m.MaximizeWindow != nil {
		m.MaximizeWindow()
	}
	m.stopCurrent()
	m.tabsMenuOpen = false
	m.tabsContextOpen = false
	m.mainMenuOpen = false
	m.projectMenuOpen = false
	m.homeVisible = false
	m.paneRoot = nil
	m.activePane = nil
	m.nextPaneID = 0
	m.focusPane = nil
	m.taskPane = nil
	m.treeScrollX = 0
	m.treeScrollGeometry = components.ScrollAreaGeometry{}
	m.generation++
	generation := m.generation
	m.taskEntry = -1
	m.busy = true
	m.status = m.tf("正在打开 %s", filepath.Base(path))
	m.diagnostic = ""
	go func() {
		opened, err := jarlensworkspace.Open(path)
		m.publish(&taskResult{kind: taskOpen, generation: generation, workspace: opened, err: err})
	}()
}

func (m *Model) startDecompile(entryID int) {
	targetID, revealName := m.decompileTarget(entryID)
	m.startDocument(targetID, revealName, taskDecompile)
	m.revealProjectEntry(entryID)
}

func (m *Model) startResource(entryID int) {
	m.startDocument(entryID, "", taskResource)
	m.revealProjectEntry(entryID)
}

func (m *Model) startDocument(entryID int, revealName string, kind taskKind) {
	m.ensurePane()
	index := m.documentIndex(entryID)
	if index < 0 {
		entry := m.workspace.Archive.Entries[entryID]
		location := m.workspace.Archive.Path
		if !strings.EqualFold(filepath.Ext(location), ".class") {
			location += "!/" + entry.Path
		}
		m.documents = append(m.documents, &document{entryID: entryID, name: entry.Name, path: entry.Path, location: location, language: languageForPath(entry.Path)})
		index = len(m.documents) - 1
	}
	m.activateDocument(index)
	m.selected = entryID
	doc := m.currentDocument()
	doc.revealName = revealName
	if m.source != "" {
		m.revealDocument(doc, revealName)
		doc.revealName = ""
		return
	}
	if doc.loading {
		return
	}

	m.stopCurrent()
	m.generation++
	generation := m.generation
	m.taskEntry = entryID
	m.taskPane = m.activePane
	doc.loading = true
	doc.diagnostic = ""
	m.source = ""
	m.sourceInfo = ""
	m.editor = text.NewEditor("")
	m.code = components.CodeEditor{}
	m.diagnostic = ""
	m.busy = true
	if kind == taskResource {
		m.status = m.tf("正在读取 %s", doc.name)
	} else {
		m.status = m.tf("正在反编译 %s", doc.name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	m.cancel = cancel
	workspace := m.workspace
	engine := m.engine
	m.syncActivePane()
	pane := m.activePane
	go func() {
		defer cancel()
		if kind == taskResource {
			content, err := workspace.TextResource(entryID)
			m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, pane: pane, result: decompile.Result{Source: content}, err: err})
			return
		}
		request, err := workspace.DecompileRequest(entryID)
		if err != nil {
			m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, pane: pane, err: err})
			return
		}
		result, err := engine.Decompile(ctx, request)
		m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, pane: pane, result: result, err: err})
	}()
}

func (m *Model) decompileTarget(entryID int) (int, string) {
	targetID := entryID
	revealName := ""
	for _, node := range m.workspace.Tree.Nodes {
		if node.Entry != entryID {
			continue
		}
		if node.Kind == archive.NodeClass {
			revealName = nestedDeclarationName(node.Name)
		}
		for node.Parent >= 0 {
			parent := m.workspace.Tree.Nodes[node.Parent]
			if parent.Kind != archive.NodeClass {
				break
			}
			targetID = parent.Entry
			if revealName == "" {
				revealName = nestedDeclarationName(parent.Name)
			}
			node = parent
		}
		break
	}
	return targetID, revealName
}

func (m *Model) documentIndex(entryID int) int {
	for index, doc := range m.documents {
		if doc.entryID == entryID {
			return index
		}
	}
	return -1
}

func (m *Model) currentDocument() *document {
	if m.active < 0 || m.active >= len(m.documents) {
		return nil
	}
	return m.documents[m.active]
}

func (m *Model) storeCurrentDocument() {
	doc := m.currentDocument()
	if doc == nil {
		return
	}
	doc.source = m.source
	doc.sourceInfo = m.sourceInfo
	doc.editor = m.editor
	doc.code = m.code
}

func (m *Model) activateDocument(index int) {
	if index < 0 || index >= len(m.documents) || index == m.active {
		return
	}
	m.storeCurrentDocument()
	doc := m.documents[index]
	m.active = index
	m.focusProject = false
	m.source = doc.source
	m.sourceInfo = doc.sourceInfo
	m.editor = doc.editor
	m.code = doc.code
	m.diagnostic = doc.diagnostic
	m.status = m.tf("已打开 %s", doc.name)
	m.syncActivePane()
}

func (m *Model) revealProjectEntry(entryID int) {
	if m.workspace == nil {
		return
	}
	m.selected = entryID
	m.treeScrollX = 0
	for _, node := range m.workspace.Tree.Nodes {
		if node.Entry != entryID {
			continue
		}
		for parent := node.Parent; parent >= 0; parent = m.workspace.Tree.Nodes[parent].Parent {
			m.expanded[parent] = true
		}
		break
	}
	visible := m.workspace.Tree.Visible(m.expanded)
	if query := m.filter.Text(); query != "" {
		visible = m.workspace.Tree.VisibleQuery(query)
	}
	for index, item := range visible {
		if item.Node.Entry != entryID {
			continue
		}
		state := m.view.ScrollState(projectTree)
		top := float32(index) * 24
		offset := top
		if state.Viewport > 0 {
			maxScroll := max(float32(len(visible))*24-state.Viewport, 0)
			offset = min(top+12-state.Viewport/2, maxScroll)
		}
		state.ScrollTo(max(0, offset))
		return
	}
}

func (m *Model) closeActiveDocument() {
	m.closeDocument(m.active)
}

func (m *Model) closeDocument(index int) {
	if index < 0 || index >= len(m.documents) {
		return
	}
	if m.busy && m.taskPane == m.activePane && m.taskEntry == m.documents[index].entryID {
		m.cancelWork()
	}
	m.documents = append(m.documents[:index], m.documents[index+1:]...)
	wasActive := m.active == index
	if index < m.active {
		m.active--
	}
	if len(m.documents) == 0 {
		m.active = -1
		m.selected = -1
		m.tabsMenuOpen = false
		m.tabsContextOpen = false
		m.clearCurrentDocument()
		m.syncActivePane()
		m.collapseEmptyPane(m.activePane)
		return
	}
	if wasActive {
		m.active = -1
		m.activateDocument(min(index, len(m.documents)-1))
	}
}

func (m *Model) closeAllDocuments() {
	for len(m.documents) > 0 {
		m.closeDocument(len(m.documents) - 1)
	}
	m.tabsMenuOpen = false
	m.status = m.t("关闭全部标签页")
}

func (m *Model) closeOtherDocuments() {
	doc := m.currentDocument()
	for index := len(m.documents) - 1; index >= 0; index-- {
		if m.documents[index] != doc {
			m.closeDocument(index)
		}
	}
	m.tabsMenuOpen = false
	m.status = m.t("关闭其他标签页")
}

func (m *Model) closeDocumentsToLeft() {
	for m.active > 0 {
		m.closeDocument(0)
	}
	m.tabsMenuOpen = false
	m.status = m.t("关闭左侧标签页")
}

func (m *Model) closeDocumentsToRight() {
	for m.active >= 0 && m.active < len(m.documents)-1 {
		m.closeDocument(m.active + 1)
	}
	m.tabsMenuOpen = false
	m.status = m.t("关闭右侧标签页")
}

func (m *Model) clearCurrentDocument() {
	m.source = ""
	m.sourceInfo = ""
	m.editor = text.NewEditor("")
	m.code = components.CodeEditor{}
	m.diagnostic = ""
}

func (m *Model) stopCurrent() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if index := paneDocumentIndex(m.taskPane, m.taskEntry); index >= 0 {
		m.taskPane.documents[index].loading = false
	}
}

func (m *Model) cancelWork() {
	if !m.busy {
		return
	}
	m.stopCurrent()
	m.generation++
	m.busy = false
	m.taskEntry = -1
	m.status = m.t("操作已取消")
	m.diagnostic = ""
}

func (m *Model) publish(result *taskResult) {
	m.mu.Lock()
	if m.pending == nil || result.generation >= m.pending.generation {
		m.pending = result
	}
	m.mu.Unlock()
	m.invalidator.Invalidate()
}

func (m *Model) wake() {
	if m.settings.Theme == "system" {
		m.applyThemePreference("system")
	}
	m.mu.Lock()
	result := m.pending
	m.pending = nil
	releases := m.pendingUpdateReleases
	m.pendingUpdateReleases = nil
	m.mu.Unlock()
	if releases != nil {
		m.applyUpdateReleasesResult(releases)
	}
	if result == nil || result.generation != m.generation {
		return
	}
	if result.kind == taskUpdate {
		m.applyUpdateResult(result)
		return
	}
	m.busy = false
	m.cancel = nil
	if result.err != nil {
		if index := paneDocumentIndex(result.pane, result.entryID); index >= 0 {
			doc := result.pane.documents[index]
			doc.loading = false
			doc.diagnostic = result.err.Error()
			if result.pane == m.activePane && index == m.active {
				m.diagnostic = doc.diagnostic
			}
		}
		m.fail(m.operationError(result.err), result.err)
		return
	}

	switch result.kind {
	case taskOpen:
		m.workspace = result.workspace
		m.recentProjects = jarlensworkspace.RecordRecent(m.recentProjects, m.workspace.Archive.Path)
		if m.RecentStore != nil {
			m.recentProjects = m.RecentStore.Record(m.workspace.Archive.Path)
		}
		m.classLinks = newClassLinkIndex(m.workspace.Archive.Entries)
		m.expanded = make(map[int]bool)
		m.documents = nil
		m.active = -1
		m.selected = -1
		m.tabsMenuOpen = false
		m.clearCurrentDocument()
		classes := 0
		for _, entry := range m.workspace.Archive.Entries {
			if entry.Kind == archive.EntryClass {
				classes++
			}
		}
		m.status = m.tf("%s · %d 个类", filepath.Base(m.workspace.Archive.Path), classes)
		m.diagnostic = ""
	case taskDecompile, taskResource:
		m.applyPaneResult(result)
	}
	m.taskEntry = -1
}

func (m *Model) saveSource() {
	doc := m.activeDocument()
	if doc == nil || doc.source == "" || m.workspace == nil {
		return
	}
	name := doc.name
	extension := filepath.Ext(name)
	if strings.EqualFold(extension, ".class") {
		name = strings.TrimSuffix(name, extension) + ".java"
		extension = ".java"
	}
	if extension == "" {
		extension = ".txt"
	}
	path, err := app.SaveFile(app.SaveFileOptions{
		Title: m.t("导出文档"), SuggestedName: name,
		DefaultExtension: extension, Filters: []app.FileFilter{{Name: m.t("文档"), Pattern: "*" + extension}},
	})
	if err != nil {
		m.fail("无法保存源码", err)
		return
	}
	if path == "" {
		return
	}
	if err := os.WriteFile(path, []byte(doc.source), 0o600); err != nil {
		m.fail("无法保存源码", err)
		return
	}
	m.status = m.tf("已导出 %s", filepath.Base(path))
}

func (m *Model) openRepository() {
	opener := m.openURL
	if opener == nil {
		opener = app.OpenURL
	}
	if err := opener(repositoryURL); err != nil {
		m.fail("无法打开 GitHub 仓库", err)
		return
	}
	m.status = m.t("正在打开 GitHub 仓库")
}

func (m *Model) close() {
	if m.statusItem != nil {
		m.statusItem.Remove()
		m.statusItem = nil
	}
	if m.stopSystemAppearance != nil {
		m.stopSystemAppearance()
		m.stopSystemAppearance = nil
	}
	m.stopCurrent()
	if m.updateCancel != nil {
		m.updateCancel()
		m.updateCancel = nil
	}
	m.cancelUpdateReleases()
	m.closeEngines()
}

func (m *Model) fail(message string, err error) {
	m.status = m.t(message)
	m.diagnostic = err.Error()
}

func (m *Model) operationError(err error) string {
	switch {
	case errors.Is(err, decompile.ErrJavaNotFound):
		return m.t("未找到 Java 运行时")
	case errors.Is(err, context.DeadlineExceeded):
		return m.t("反编译超时")
	case errors.Is(err, context.Canceled):
		return m.t("操作已取消")
	case errors.Is(err, archive.ErrUnsupportedFormat):
		return m.t("仅支持 JAR 和 Class 文件")
	case errors.Is(err, archive.ErrInvalidClass):
		return m.t("所选文件不是有效的 Java 类文件")
	default:
		return m.t("操作失败")
	}
}
