package ui

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/xddcode/jarlens/internal/engine/vineflower"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

const repositoryURL = "https://github.com/xddcode/jarlens"

type taskKind uint8

const (
	taskOpen taskKind = iota
	taskDecompile
	taskResource
)

type taskResult struct {
	kind       taskKind
	generation uint64
	entryID    int
	workspace  *jarlensworkspace.Workspace
	result     decompile.Result
	err        error
}

type document struct {
	entryID    int
	name       string
	path       string
	language   string
	source     string
	sourceInfo string
	diagnostic string
	revealName string
	loading    bool
	editor     *text.Editor
	code       components.CodeEditor
}

type Model struct {
	Theme Appearance

	view       flex.View
	tooltip    components.HoverTooltip
	code       components.CodeEditor
	codeFace   vivaui.MonoFace
	split      flex.SplitterState
	editor     *text.Editor
	engine     decompile.Engine
	workspace  *jarlensworkspace.Workspace
	expanded   map[int]bool
	documents  []*document
	active     int
	selected   int
	source     string
	sourceInfo string
	busy       bool
	status     string
	diagnostic string

	invalidator *app.Invalidator
	mu          sync.Mutex
	pending     *taskResult
	generation  uint64
	taskEntry   int
	cancel      context.CancelFunc
	openURL     func(string) error
}

func (m *Model) initialize() {
	configureFonts()
	m.codeFace = monoFace
	if m.split.Pos == 0 {
		m.split.Pos = 280
	}
	if m.editor == nil {
		m.editor = text.NewEditor("")
	}
	if m.engine == nil {
		m.engine = vineflower.New()
	}
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
}

func (m *Model) OpenPath(path string) {
	m.startOpen(path)
}

func (m *Model) chooseArchive() {
	paths, err := app.OpenFiles(app.OpenFileOptions{
		Title:   "Open Java archive",
		Filters: []app.FileFilter{{Name: "Java archives and classes", Pattern: "*.jar;*.class"}},
	})
	if err != nil {
		m.fail("Could not open file dialog", err)
		return
	}
	if len(paths) > 0 {
		m.startOpen(paths[0])
	}
}

func (m *Model) startOpen(path string) {
	m.stopCurrent()
	m.generation++
	generation := m.generation
	m.taskEntry = -1
	m.busy = true
	m.status = "Opening " + filepath.Base(path)
	m.diagnostic = ""
	go func() {
		opened, err := jarlensworkspace.Open(path)
		m.publish(&taskResult{kind: taskOpen, generation: generation, workspace: opened, err: err})
	}()
}

func (m *Model) startDecompile(entryID int) {
	targetID, revealName := m.decompileTarget(entryID)
	m.startDocument(targetID, revealName, taskDecompile)
	m.selected = entryID
}

func (m *Model) startResource(entryID int) {
	m.startDocument(entryID, "", taskResource)
}

func (m *Model) startDocument(entryID int, revealName string, kind taskKind) {
	index := m.documentIndex(entryID)
	if index < 0 {
		entry := m.workspace.Archive.Entries[entryID]
		m.documents = append(m.documents, &document{entryID: entryID, name: entry.Name, path: entry.Path, language: languageForPath(entry.Path)})
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
	doc.loading = true
	doc.diagnostic = ""
	m.source = ""
	m.sourceInfo = ""
	m.editor = text.NewEditor("")
	m.code = components.CodeEditor{}
	m.diagnostic = ""
	m.busy = true
	if kind == taskResource {
		m.status = "Reading " + doc.name
	} else {
		m.status = "Decompiling " + doc.name
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	m.cancel = cancel
	workspace := m.workspace
	engine := m.engine
	go func() {
		defer cancel()
		if kind == taskResource {
			content, err := workspace.TextResource(entryID)
			m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, result: decompile.Result{Source: content}, err: err})
			return
		}
		request, err := workspace.DecompileRequest(entryID)
		if err != nil {
			m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, err: err})
			return
		}
		result, err := engine.Decompile(ctx, request)
		m.publish(&taskResult{kind: kind, generation: generation, entryID: entryID, result: result, err: err})
	}()
}

func (m *Model) decompileTarget(entryID int) (int, string) {
	targetID := entryID
	revealName := ""
	for _, node := range m.workspace.Tree.Nodes {
		if node.Entry != entryID {
			continue
		}
		for node.Parent >= 0 {
			parent := m.workspace.Tree.Nodes[node.Parent]
			if parent.Kind != archive.NodeClass {
				break
			}
			targetID = parent.Entry
			if revealName == "" {
				revealName = strings.TrimSuffix(node.Name, filepath.Ext(node.Name))
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
	m.selected = doc.entryID
	m.source = doc.source
	m.sourceInfo = doc.sourceInfo
	m.editor = doc.editor
	m.code = doc.code
	m.diagnostic = doc.diagnostic
}

func (m *Model) closeActiveDocument() {
	index := m.active
	if index < 0 || index >= len(m.documents) {
		return
	}
	if m.busy && m.taskEntry == m.documents[index].entryID {
		m.cancelWork()
	}
	m.documents = append(m.documents[:index], m.documents[index+1:]...)
	m.active = -1
	if len(m.documents) == 0 {
		m.selected = -1
		m.clearCurrentDocument()
		return
	}
	m.activateDocument(min(index, len(m.documents)-1))
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
}

func (m *Model) cancelWork() {
	if !m.busy {
		return
	}
	m.stopCurrent()
	m.generation++
	m.busy = false
	if index := m.documentIndex(m.taskEntry); index >= 0 {
		m.documents[index].loading = false
	}
	m.taskEntry = -1
	m.status = "Operation cancelled"
	m.diagnostic = ""
}

func (m *Model) publish(result *taskResult) {
	m.mu.Lock()
	m.pending = result
	m.mu.Unlock()
	m.invalidator.Invalidate()
}

func (m *Model) wake() {
	m.mu.Lock()
	result := m.pending
	m.pending = nil
	m.mu.Unlock()
	if result == nil || result.generation != m.generation {
		return
	}
	m.busy = false
	m.cancel = nil
	if result.err != nil {
		if index := m.documentIndex(result.entryID); index >= 0 {
			doc := m.documents[index]
			doc.loading = false
			doc.diagnostic = result.err.Error()
			if index == m.active {
				m.diagnostic = doc.diagnostic
			}
		}
		m.fail(operationError(result.err), result.err)
		return
	}

	switch result.kind {
	case taskOpen:
		m.workspace = result.workspace
		m.expanded = make(map[int]bool)
		m.documents = nil
		m.active = -1
		m.selected = -1
		m.clearCurrentDocument()
		classes := 0
		for _, entry := range m.workspace.Archive.Entries {
			if entry.Kind == archive.EntryClass {
				classes++
			}
		}
		m.status = fmt.Sprintf("%s · %d classes", filepath.Base(m.workspace.Archive.Path), classes)
		m.diagnostic = ""
	case taskDecompile, taskResource:
		index := m.documentIndex(result.entryID)
		if index < 0 {
			return
		}
		doc := m.documents[index]
		doc.loading = false
		doc.source = result.result.Source
		doc.sourceInfo = "Decompiled by " + result.result.Engine + " " + result.result.EngineVersion
		doc.diagnostic = ""
		doc.editor = text.NewEditor(doc.source)
		doc.code = components.CodeEditor{}
		if index == m.active {
			m.source = doc.source
			m.sourceInfo = doc.sourceInfo
			m.editor = doc.editor
			m.code = doc.code
			m.diagnostic = ""
		}
		if doc.revealName != "" {
			m.revealDocument(doc, doc.revealName)
			doc.revealName = ""
		}
		if result.kind == taskResource {
			doc.sourceInfo = "Text resource · UTF-8"
			if index == m.active {
				m.sourceInfo = doc.sourceInfo
			}
			m.status = "Opened " + doc.name
		} else {
			m.status = result.result.Engine + " " + result.result.EngineVersion
		}
	}
	m.taskEntry = -1
}

func (m *Model) saveSource() {
	if m.source == "" || m.workspace == nil || m.selected < 0 {
		return
	}
	name := m.currentDocument().name
	extension := filepath.Ext(name)
	if strings.EqualFold(extension, ".class") {
		name = strings.TrimSuffix(name, extension) + ".java"
		extension = ".java"
	}
	if extension == "" {
		extension = ".txt"
	}
	path, err := app.SaveFile(app.SaveFileOptions{
		Title: "Export document", SuggestedName: name,
		DefaultExtension: extension, Filters: []app.FileFilter{{Name: "Document", Pattern: "*" + extension}},
	})
	if err != nil {
		m.fail("Could not save source", err)
		return
	}
	if path == "" {
		return
	}
	if err := os.WriteFile(path, []byte(m.source), 0o600); err != nil {
		m.fail("Could not save source", err)
		return
	}
	m.status = "Exported " + filepath.Base(path)
}

func (m *Model) openRepository() {
	opener := m.openURL
	if opener == nil {
		opener = app.OpenURL
	}
	if err := opener(repositoryURL); err != nil {
		m.fail("Could not open GitHub repository", err)
		return
	}
	m.status = "Opening GitHub repository"
}

func (m *Model) close() {
	m.stopCurrent()
	if closer, ok := m.engine.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func (m *Model) fail(message string, err error) {
	m.status = message
	m.diagnostic = err.Error()
}

func operationError(err error) string {
	switch {
	case errors.Is(err, decompile.ErrJavaNotFound):
		return "Java runtime not found"
	case errors.Is(err, context.DeadlineExceeded):
		return "Decompilation timed out"
	case errors.Is(err, context.Canceled):
		return "Operation cancelled"
	case errors.Is(err, archive.ErrUnsupportedFormat):
		return "Only JAR and Class files are supported"
	case errors.Is(err, archive.ErrInvalidClass):
		return "The selected file is not a valid Java class"
	default:
		return "Operation failed"
	}
}
