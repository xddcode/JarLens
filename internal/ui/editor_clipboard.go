package ui

import (
	"github.com/vivaui/viva/app"
	"unicode/utf8"

	"github.com/vivaui/viva/text"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) selectAllSource() bool {
	if m.editor == nil || m.source == "" {
		return false
	}
	m.editor.SetSelection(text.Range{Start: m.editor.RuneCount(), End: 0})
	return true
}

func (m *Model) copySource() bool {
	if m.editor == nil || m.source == "" {
		return false
	}
	selection := m.editor.Selection()
	if selection.Start == selection.End {
		return false
	}
	start := m.editor.ByteOffset(selection.Start)
	end := m.editor.ByteOffset(selection.End)
	if err := app.SetClipboardText(m.source[start:end]); err != nil {
		m.status = m.t("无法复制到剪贴板")
		return false
	}
	m.status = m.t("已复制所选内容")
	return true
}

func (m *Model) sourceSelectionPointer(event vivaui.PointerEvent, pane *editorPane) (vivaui.EventResult, bool) {
	if pane == nil || pane != m.activePane || event.Target != m.paneSourceID(pane) || m.editor == nil {
		return vivaui.EventResult{}, false
	}
	switch event.Kind {
	case vivaui.PointerDown:
		offset, ok := m.code.OffsetAt(vivaui.Point{X: event.X, Y: event.Y})
		if !ok {
			return vivaui.EventResult{}, false
		}
		m.moveSourceCaret(offset, false)
		m.selectingSource = true
		return vivaui.EventResult{Handled: true, Repaint: true}, true
	case vivaui.PointerMove:
		if !m.selectingSource {
			return vivaui.EventResult{}, false
		}
		offset, ok := m.code.OffsetAt(vivaui.Point{X: event.X, Y: event.Y})
		if ok {
			m.moveSourceCaret(offset, true)
		}
		return vivaui.EventResult{Handled: true, Repaint: true}, true
	case vivaui.PointerUp:
		if !m.selectingSource {
			return vivaui.EventResult{}, false
		}
		m.selectingSource = false
		if offset, ok := m.code.OffsetAt(vivaui.Point{X: event.X, Y: event.Y}); ok {
			m.moveSourceCaret(offset, true)
		}
		return vivaui.EventResult{Handled: true, Repaint: true}, true
	default:
		return vivaui.EventResult{}, false
	}
}

func (m *Model) moveSourceCaret(offset int, extend bool) {
	value := m.editor.String()
	m.editor.MoveTo(utf8.RuneCountInString(value[:offset]), extend)
}
