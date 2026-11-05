package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
)

type taskResult struct {
	kind       taskKind
	generation uint64
	entryID    int
	workspace  *jarlensworkspace.Workspace
	result     decompile.Result
	err        error
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
	if m.selected == 0 && m.workspace == nil {
		m.selected = -1
	}
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
	m.busy = true
	m.status = "Opening " + filepath.Base(path)
	m.diagnostic = ""
	go func() {
		opened, err := jarlensworkspace.Open(path)
		m.publish(&taskResult{kind: taskOpen, generation: generation, workspace: opened, err: err})
	}()
}

func (m *Model) startDecompile(entryID int) {
	m.stopCurrent()
	m.generation++
	generation := m.generation
	m.selected = entryID
	m.source = ""
	m.editor = text.NewEditor("")
	m.diagnostic = ""
	m.busy = true
	m.status = "Decompiling " + m.workspace.Archive.Entries[entryID].Name
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	m.cancel = cancel
	workspace := m.workspace
	engine := m.engine
	go func() {
		defer cancel()
		request, err := workspace.DecompileRequest(entryID)
		if err != nil {
			m.publish(&taskResult{kind: taskDecompile, generation: generation, entryID: entryID, err: err})
			return
		}
		result, err := engine.Decompile(ctx, request)
		m.publish(&taskResult{kind: taskDecompile, generation: generation, entryID: entryID, result: result, err: err})
	}()
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
		m.fail(operationError(result.err), result.err)
		return
	}

	switch result.kind {
	case taskOpen:
		m.workspace = result.workspace
		m.expanded = make(map[int]bool)
		m.selected = -1
		m.source = ""
		m.sourceInfo = ""
		m.editor = text.NewEditor("")
		classes := 0
		for _, entry := range m.workspace.Archive.Entries {
			if entry.Kind == archive.EntryClass {
				classes++
			}
		}
		m.status = fmt.Sprintf("%s · %d classes", filepath.Base(m.workspace.Archive.Path), classes)
		m.diagnostic = ""
	case taskDecompile:
		m.source = result.result.Source
		m.editor = text.NewEditor(m.source)
		m.status = result.result.Engine + " " + result.result.EngineVersion
		m.sourceInfo = "Decompiled by " + result.result.Engine + " " + result.result.EngineVersion
		m.diagnostic = ""
	}
}

func (m *Model) saveSource() {
	if m.source == "" || m.workspace == nil || m.selected < 0 {
		return
	}
	name := m.workspace.Archive.Entries[m.selected].Name
	path, err := app.SaveFile(app.SaveFileOptions{
		Title: "Export decompiled source", SuggestedName: name + ".java",
		DefaultExtension: ".java", Filters: []app.FileFilter{{Name: "Java source", Pattern: "*.java"}},
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
