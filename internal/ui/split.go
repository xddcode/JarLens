package ui

import (
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/text"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) newPane() *editorPane {
	pane := &editorPane{id: m.nextPaneID, active: -1, editor: text.NewEditor("")}
	m.nextPaneID++
	return pane
}

func (m *Model) ensurePane() {
	if m.paneRoot != nil {
		return
	}
	pane := m.newPane()
	pane.documents = m.documents
	pane.active = m.active
	pane.editor = m.editor
	pane.code = m.code
	m.paneRoot = pane
	m.activePane = pane
}

func (pane *editorPane) leaf() bool {
	return pane != nil && pane.first == nil
}

func (pane *editorPane) current() *document {
	if pane == nil || pane.active < 0 || pane.active >= len(pane.documents) {
		return nil
	}
	return pane.documents[pane.active]
}

func (m *Model) syncActivePane() {
	if m.activePane == nil || !m.activePane.leaf() {
		return
	}
	m.storeCurrentDocument()
	m.activePane.documents = m.documents
	m.activePane.active = m.active
	m.activePane.editor = m.editor
	m.activePane.code = m.code
}

func (m *Model) loadPane(pane *editorPane) {
	m.activePane = pane
	m.documents = pane.documents
	m.active = pane.active
	m.editor = pane.editor
	m.code = pane.code
	doc := m.currentDocument()
	if doc == nil {
		m.clearCurrentDocument()
		return
	}
	m.source = doc.source
	m.sourceInfo = doc.sourceInfo
	m.diagnostic = doc.diagnostic
	m.selected = doc.entryID
}

func (m *Model) activatePane(pane *editorPane) {
	if pane == nil || !pane.leaf() {
		return
	}
	if m.activePane == pane {
		return
	}
	m.syncActivePane()
	m.loadPane(pane)
}

func (m *Model) activeDocument() *document {
	return m.currentDocument()
}

func paneDocumentIndex(pane *editorPane, entryID int) int {
	if pane == nil {
		return -1
	}
	for index, doc := range pane.documents {
		if doc.entryID == entryID {
			return index
		}
	}
	return -1
}

func (m *Model) splitDocument(index int, direction splitDirection) {
	pane := m.tabsContextPane
	if pane == nil {
		pane = m.activePane
	}
	m.activatePane(pane)
	if index < 0 || index >= len(m.documents) || m.documents[index].source == "" {
		return
	}
	m.syncActivePane()
	first := &editorPane{id: pane.id, editor: pane.editor}
	first.parent = pane
	first.documents = pane.documents
	first.active = pane.active
	first.editor = pane.editor
	first.code = pane.code
	source := m.documents[index]
	clone := *source
	clone.editor = text.NewEditor(source.source)
	clone.editor.SetSelection(source.editor.Selection())
	clone.code = components.CodeEditor{}
	clone.pendingCenter = false
	second := m.newPane()
	second.parent = pane
	second.documents = []*document{&clone}
	second.active = 0
	second.editor = clone.editor
	second.code = clone.code
	pane.axis = direction
	pane.splitter.Pos = 240
	if direction == splitRight && pane.width > 0 {
		pane.splitter.Pos = pane.width / 2
	}
	if direction == splitDown {
		pane.splitter.Pos = 200
		if pane.height > 0 {
			pane.splitter.Pos = pane.height/2 + 38
		}
	}
	pane.first = first
	pane.second = second
	pane.documents = nil
	pane.active = -1
	m.loadPane(second)
	m.focusPane = second
	m.tabsContextOpen = false
	if direction == splitDown {
		m.status = m.tf("已向下拆分 %s", source.name)
	} else {
		m.status = m.tf("已向右拆分 %s", source.name)
	}
}

func (m *Model) collapseEmptyPane(pane *editorPane) {
	if pane == nil || pane.parent == nil {
		return
	}
	parent := pane.parent
	remaining := parent.first
	if remaining == pane {
		remaining = parent.second
	}
	grandparent := parent.parent
	*parent = *remaining
	parent.parent = grandparent
	if parent.first != nil {
		parent.first.parent = parent
		parent.second.parent = parent
	}
	m.loadPane(parent)
	m.focusPane = parent
}

func (m *Model) walkPanes(visit func(*editorPane)) {
	var walk func(*editorPane)
	walk = func(pane *editorPane) {
		if pane == nil {
			return
		}
		if pane.leaf() {
			visit(pane)
			return
		}
		walk(pane.first)
		walk(pane.second)
	}
	walk(m.paneRoot)
}

func (m *Model) paneSourceID(pane *editorPane) vivaui.ID {
	if pane.id == 0 {
		return sourceEditor
	}
	if pane.id == 1 {
		return splitSourceEditor
	}
	return splitSourceEditors.At(pane.id)
}

func (m *Model) paneScrollbarID(pane *editorPane) vivaui.ID {
	if pane.id == 0 {
		return sourceScrollbar
	}
	if pane.id == 1 {
		return splitSourceScrollbar
	}
	return splitSourceScrollbars.At(pane.id)
}

func (m *Model) paneHorizontalScrollbarID(pane *editorPane) vivaui.ID {
	if pane.id == 0 {
		return sourceHorizontalScrollbar
	}
	if pane.id == 1 {
		return splitSourceHorizontalScrollbar
	}
	return splitSourceHorizontalScrollbars.At(pane.id)
}

func paneTabID(pane *editorPane, index int) vivaui.ID {
	if pane.id == 0 {
		return editorTabs.At(index)
	}
	if pane.id == 1 {
		return splitEditorTabs.At(index)
	}
	return extraEditorTabs.At(pane.id*1000 + index)
}

func paneCloseID(pane *editorPane, index int) vivaui.ID {
	if pane.id == 0 {
		return editorTabClose.At(index)
	}
	if pane.id == 1 {
		return splitEditorTabClose.At(index)
	}
	return extraEditorTabClose.At(pane.id*1000 + index)
}

func paneMenuID(pane *editorPane) vivaui.ID {
	if pane.id == 0 {
		return editorTabsMenuOpen
	}
	if pane.id == 1 {
		return splitEditorTabsMenuOpen
	}
	return extraEditorTabMenus.At(pane.id)
}

func (m *Model) paneForSource(id vivaui.ID) *editorPane {
	var found *editorPane
	m.walkPanes(func(pane *editorPane) {
		sourceID := m.paneSourceID(pane)
		if id == sourceID || id == m.paneScrollbarID(pane) || id == m.paneHorizontalScrollbarID(pane) {
			found = pane
		}
	})
	return found
}

func (m *Model) paneForSplitter(id vivaui.ID) *editorPane {
	index, ok := id.In(editorPaneSplitters)
	if !ok {
		return nil
	}
	var find func(*editorPane) *editorPane
	find = func(pane *editorPane) *editorPane {
		if pane == nil {
			return nil
		}
		if !pane.leaf() && pane.id == index {
			return pane
		}
		if pane.leaf() {
			return nil
		}
		if found := find(pane.first); found != nil {
			return found
		}
		return find(pane.second)
	}
	return find(m.paneRoot)
}

func (m *Model) applyPaneResult(result *taskResult) {
	pane := result.pane
	index := paneDocumentIndex(pane, result.entryID)
	if index < 0 {
		return
	}
	doc := pane.documents[index]
	doc.loading = false
	doc.source = reindentSource(result.result.Source, doc.language, sourceIndentWidth(doc.language), m.settings.TabWidth)
	doc.sourceInfo = m.t("反编译引擎：") + result.result.Engine + " " + result.result.EngineVersion
	doc.diagnostic = ""
	doc.editor = text.NewEditor(doc.source)
	doc.code = components.CodeEditor{}
	doc.links = buildClassLinks(doc.source, doc.path, m.classLinks)
	if index == pane.active {
		pane.editor = doc.editor
		pane.code = doc.code
	}
	if pane == m.activePane {
		m.editor = pane.editor
		m.code = pane.code
		m.source = doc.source
		m.sourceInfo = doc.sourceInfo
		m.diagnostic = ""
	}
	if doc.revealName != "" {
		m.revealDocument(doc, doc.revealName)
		doc.revealName = ""
	}
	if result.kind == taskResource {
		doc.sourceInfo = m.t("文本资源 · UTF-8")
		if pane == m.activePane {
			m.sourceInfo = doc.sourceInfo
		}
		m.status = m.tf("已打开 %s", doc.name)
	} else {
		m.status = result.result.Engine + " " + result.result.EngineVersion
	}
}
