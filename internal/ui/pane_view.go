package ui

import (
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
)

func (m *Model) paintEditorPanes(f *flex.Frame, theme vivaui.Theme) (map[*editorPane]flex.Box, map[*editorPane]flex.Box) {
	m.ensurePane()
	sources := make(map[*editorPane]flex.Box)
	menus := make(map[*editorPane]flex.Box)
	var paint func(*editorPane)
	paint = func(pane *editorPane) {
		if !pane.leaf() {
			if pane.axis == splitRight {
				f.Row().Grow(1).Children(func() {
					f.Column().Width(pane.splitter.Pos).Children(func() { paint(pane.first) })
					f.Splitter(editorPaneSplitters.At(pane.id), &pane.splitter).Width(6).Mins(160, 160).Background(theme.Card)
					f.Column().Grow(1).Children(func() { paint(pane.second) })
				})
			} else {
				f.Column().Grow(1).Children(func() {
					f.Column().Height(pane.splitter.Pos).Children(func() { paint(pane.first) })
					f.Splitter(editorPaneSplitters.At(pane.id), &pane.splitter).Height(6).Mins(120, 120).Background(theme.Card)
					f.Column().Grow(1).Children(func() { paint(pane.second) })
				})
			}
			return
		}
		f.Column().Grow(1).Background(theme.Background).Children(func() {
			menus[pane] = m.paintEditorHeaderForPane(f, theme, pane)
			if pane == m.activePane && m.findOpen {
				m.paintFindBar(f, theme)
				f.Divider()
			}
			doc := pane.current()
			if pane == m.activePane {
				doc = m.currentDocument()
			}
			switch {
			case doc != nil && doc.loading:
				f.Column().Grow(1).Center().Gap(12).Children(func() {
					f.Spinner()
					f.Text(m.status).FontSize(13).Muted()
				})
			case doc != nil && doc.source != "":
				f.Row().Height(36).PadX(12).Gap(8).Items(flex.Center).Background(editorBannerBackground(m.Theme)).Children(func() {
					f.Icon(infoGlyph).Size(14, 14).Color(theme.MutedForeground)
					f.Text(doc.sourceInfo).FontSize(12).Color(theme.AccentForeground).Ellipsis()
				})
				f.Divider()
				sources[pane] = f.Slot().Grow(1)
			case doc != nil && doc.diagnostic != "":
				f.Column().Grow(1).Justify(flex.Center).Pad(48).Gap(8).Children(func() {
					f.Text("无法显示内容").FontSize(18).Bold()
					f.Paragraph(doc.diagnostic, 4).Muted()
				})
			case len(paneDocuments(m, pane)) == 0:
				m.paintWelcome(f, theme)
			default:
				f.Column().Grow(1).Center().Children(func() {
					f.Text("没有可显示的源代码").FontSize(16).Muted()
				})
			}
		})
	}
	paint(m.paneRoot)
	return sources, menus
}

func paneDocuments(m *Model, pane *editorPane) []*document {
	if pane == m.activePane {
		return m.documents
	}
	return pane.documents
}

func (m *Model) paintPaneSources(canvas *vivaui.Canvas, theme vivaui.Theme, sources map[*editorPane]flex.Box) {
	for pane, slot := range sources {
		doc := pane.current()
		editor := pane.editor
		code := &pane.code
		pointer := vivaui.Point{}
		inside := false
		if pane == m.activePane {
			doc = m.currentDocument()
			editor = m.editor
			code = &m.code
			pointer = m.sourcePointer
			inside = m.sourcePointerInside
		}
		if doc == nil {
			continue
		}
		pane.width = slot.Rect().Width
		pane.height = slot.Rect().Height
		m.paintSourceView(canvas, theme, slot.Rect(), doc, editor, code,
			m.paneSourceID(pane), m.paneScrollbarID(pane), m.paneHorizontalScrollbarID(pane),
			pane == m.activePane && m.findOpen, pointer, inside)
	}
}

func (m *Model) paintEditorHeaderForPane(f *flex.Frame, theme vivaui.Theme, pane *editorPane) flex.Box {
	documents := paneDocuments(m, pane)
	active := pane.active
	if pane == m.activePane {
		active = m.active
	}
	if len(documents) == 0 {
		return flex.Box{}
	}
	selectedBackground, selectedBorder, hoverBackground := editorTabColors(m.Theme)
	var more flex.Box
	f.Row().Height(40).PadX(4).Gap(2).Items(flex.Center).Children(func() {
		for index, doc := range documents {
			tabID := paneTabID(pane, index)
			closeID := paneCloseID(pane, index)
			selected := index == active
			hovered := f.Canvas().Hovered() == tabID || f.Canvas().Hovered() == closeID
			tab := f.Row().Height(28).MinWidth(40).MaxWidth(300).Radius(6).PadSides(0, 0, 0, 10).Gap(6).Items(flex.Center).
				Hover(flex.Look{}.Background(hoverBackground)).
				HitWith(tabID, doc.name, vivaui.HitOptions{Role: vivaui.RoleTab, Selected: selected, Hint: doc.location, Cursor: vivaui.PointerCursorHand})
			if selected {
				tab = tab.Background(selectedBackground).Stroke(1, selectedBorder)
			} else if hovered {
				tab = tab.Background(hoverBackground)
			}
			tab.Children(func() {
				kind := archive.NodeResource
				if doc.language == "java" {
					kind = archive.NodeClass
				}
				paintFileIcon(f, doc.path, kind, m.Theme == AppearanceDark, 14)
				f.Text(doc.name).FontSize(13).MaxWidth(260).Ellipsis().Color(theme.AccentForeground)
				closeHovered := f.Canvas().Hovered() == closeID
				closePressed := f.Canvas().Pressed() == closeID
				close := f.Row().Size(20, 20).MarginSides(0, 4, 0, 0).Center().Radius(4).
					HitWith(closeID, "关闭 "+doc.name, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "关闭 " + doc.name + " (Ctrl+W)", Cursor: vivaui.PointerCursorHand})
				if f.Canvas().FocusVisible() == closeID {
					close = close.Stroke(2, theme.Ring)
				}
				close.Children(func() {
					if selected || hovered || f.Canvas().Focused() == closeID {
						closeColor := theme.MutedForeground
						if closeHovered || closePressed {
							closeColor = theme.PrimaryForeground
						}
						button := f.Row().Size(16, 16).Radius(8).Center()
						if closeHovered {
							button = button.Background(theme.Primary)
						}
						if closePressed {
							button = button.Background(theme.Primary.Mix(theme.Background, 0.16))
						}
						button.Children(func() {
							f.Icon(cancelGlyph).Size(11, 11).Color(closeColor)
						})
					}
				})
				if close.Clicked() {
					m.activatePane(pane)
					f.Canvas().AfterPaint(func() { m.closeDocument(index) })
				}
			})
			if event, ok := f.Canvas().PointerOn(tabID); ok && event.Kind == vivaui.PointerDown && event.Button == vivaui.PointerButtonSecondary {
				m.activatePane(pane)
				m.activateDocument(index)
				m.tabsContextTab = index
				m.tabsContextPane = pane
				m.tabsContextPoint = vivaui.Point{X: event.X, Y: event.Y}
				m.tabsContextOpen = true
				m.tabsMenuOpen = false
				f.Canvas().Consume()
			}
			if tab.Clicked() {
				m.activatePane(pane)
				m.activateDocument(index)
			}
		}
		f.Spacer()
		menuOpen := m.tabsMenuOpen && m.tabsMenuPane == pane
		moreHover := theme.Hover
		moreColor := tabActionsColor(m.Theme)
		if menuOpen {
			moreHover = toolbarPressedColor(m.Theme)
		}
		more = f.Row().Size(28, 28).Center().Radius(6).Hover(flex.Look{}.Background(moreHover)).
			HitWith(paneMenuID(pane), "标签页操作", vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "标签页操作", Expanded: menuOpen, Cursor: vivaui.PointerCursorHand})
		if menuOpen {
			more = more.Background(moreHover)
		}
		more.Children(func() {
			f.Icon(moreVerticalGlyph).Size(16, 16).Color(moreColor)
		})
		if more.Clicked() {
			m.tabsMenuOpen = !(m.tabsMenuOpen && m.tabsMenuPane == pane)
			m.tabsMenuPane = pane
			m.activatePane(pane)
		}
	})
	f.Divider()
	return more
}
