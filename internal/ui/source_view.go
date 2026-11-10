package ui

import (
	"time"

	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/text"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) paintSourceView(canvas *vivaui.Canvas, theme vivaui.Theme, bounds vivaui.Rect, doc *document, editor *text.Editor, code *components.CodeEditor, id, scrollbarID, horizontalScrollbarID vivaui.ID, findOpen bool, pointer vivaui.Point, pointerInside bool) {
	if doc == nil || editor == nil {
		return
	}
	value := editor.String()
	selection := editor.Selection()
	focused := canvas.Focused() == id
	showCaret := true
	if focused {
		const visible = 530 * time.Millisecond
		phase := canvas.Now() % time.Second
		showCaret = phase < visible
		next := visible - phase
		if !showCaret {
			next = time.Second - phase
		}
		canvas.RequestFrameIn(next)
	}
	snapshot := components.CodeEditorSnapshot{
		ID: id, Label: doc.name, Value: value,
		ReadOnly: true, Focused: focused, ShowCaret: showCaret,
		Caret:          editor.ByteOffset(editor.Caret()),
		SelectionStart: editor.ByteOffset(selection.Start), SelectionEnd: editor.ByteOffset(selection.End),
		FontSize: float32(m.settings.FontSize), MonoFace: m.codeFace, TabWidth: m.settings.TabWidth,
		LineNumbers: m.settings.ShowLineNumbers, CurrentLine: m.settings.HighlightCurrentLine, MatchBrackets: true,
		ShowIndentGuides: m.settings.ShowIndentGuides, InvisibleColor: theme.MutedForeground,
		Language: doc.language, ScrollbarID: scrollbarID, HorizontalScrollbarID: horizontalScrollbarID,
	}
	if findOpen && len(m.findMatches) > 0 {
		snapshot.Marks = m.findMatches
		snapshot.CurrentMark = m.findMatches[m.findIndex]
	}
	if composition, ok := editor.Composition(); ok {
		snapshot.CompositionStart = editor.ByteOffset(composition.Start)
		snapshot.CompositionEnd = editor.ByteOffset(composition.End)
	}
	snapshot.Highlight = func(language, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
		dst = highlightDocumentLine(m.Theme, language, line, dst)
		for _, link := range buildClassLinks(line, doc.path, m.classLinks) {
			dst = append(dst, components.CodeEditorRun{Start: link.start, End: link.end, Color: linkColor(m.Theme)})
		}
		return dst
	}
	sourceTheme := theme
	sourceTheme.Primary = theme.Input
	sourceTheme.Foreground, sourceTheme.MutedForeground, sourceTheme.Selection, sourceTheme.SelectionForeground = editorColors(m.Theme)
	code.Paint(canvas, sourceTheme, bounds, snapshot)
	if doc.pendingCenter {
		if target, ok := code.YAtOffset(doc.pendingCenterOffset); ok {
			code.SetScrollY(target - bounds.Height/2)
			doc.pendingCenter = false
			canvas.RequestFrame()
		}
	}
	underlineClassLinks(canvas, doc.path, m.classLinks)
	if pointerInside && canvas.Hovered() == id {
		if offset, ok := code.OffsetAt(pointer); ok {
			if _, linked := classLinkAt(doc.links, offset); linked {
				canvas.HitWithOptions(id, bounds, doc.name, vivaui.HitOptions{Role: vivaui.RoleTextField, Cursor: vivaui.PointerCursorHand, Value: value})
			}
		}
	}
}
