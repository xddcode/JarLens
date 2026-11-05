package ui

import (
	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) paintSource(canvas *vivaui.Canvas, theme vivaui.Theme, bounds vivaui.Rect) {
	doc := m.currentDocument()
	value := m.editor.String()
	selection := m.editor.Selection()
	snapshot := components.CodeEditorSnapshot{
		ID: sourceEditor, Label: doc.name, Value: value,
		ReadOnly: true, Focused: canvas.Focused() == sourceEditor, ShowCaret: true,
		Caret:          m.editor.ByteOffset(m.editor.Caret()),
		SelectionStart: m.editor.ByteOffset(selection.Start), SelectionEnd: m.editor.ByteOffset(selection.End),
		FontSize: 13, MonoFace: m.codeFace, TabWidth: 4,
		LineNumbers: true, CurrentLine: true, MatchBrackets: true,
		Language: doc.language, ScrollbarID: sourceScrollbar, HorizontalScrollbarID: sourceHorizontalScrollbar,
	}
	if composition, ok := m.editor.Composition(); ok {
		snapshot.CompositionStart = m.editor.ByteOffset(composition.Start)
		snapshot.CompositionEnd = m.editor.ByteOffset(composition.End)
	}
	snapshot.Highlight = func(language, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
		return highlightDocumentLine(m.Theme, language, line, dst)
	}
	sourceTheme := theme
	sourceTheme.Primary = theme.Input
	m.code.Paint(canvas, sourceTheme, bounds, snapshot)
}
