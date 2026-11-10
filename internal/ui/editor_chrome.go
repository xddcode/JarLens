package ui

import (
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) paintFindBar(f *flex.Frame, theme vivaui.Theme) {
	f.Row().Height(40).PadSides(6, 8, 6, 10).Gap(6).Items(flex.Center).Background(theme.CardElevated).Children(func() {
		if f.Field(&m.findField, sourceFind, m.t("在源码中查找"), m.t("查找")).Grow(1).MaxWidth(420).Submitted() {
			m.findNext(1)
		}
		f.Text(m.findStatus()).Width(64).FontSize(13).Muted()
		if smallIconButton(f, findPrevious, m.t("上一个匹配项"), arrowUpGlyph, theme, true) {
			m.findNext(-1)
		}
		if smallIconButton(f, findNext, m.t("下一个匹配项"), arrowDownGlyph, theme, true) {
			m.findNext(1)
		}
		f.Spacer()
		if smallIconButton(f, findClose, m.t("关闭查找"), cancelGlyph, theme, true) {
			m.findOpen = false
		}
	})
}

func (m *Model) paintGoLineDialog(f *flex.Frame, theme vivaui.Theme) {
	dialog := f.Dialog(flex.DialogSpec{
		Title: m.t("转到行"), Width: 360,
		Scope: goLineDialog, DismissID: goLineClose, CloseID: goLineClose, InitialFocusID: goLineInput,
	})
	dialog.Body(func() {
		f.Row().Gap(8).Items(flex.Center).Children(func() {
			if f.Field(&m.goLineField, goLineInput, m.t("行号"), m.t("行号")).Grow(1).Submitted() {
				if m.goToLine(m.goLineField.Text()) {
					m.goLineOpen = false
				} else {
					m.status = m.t("无效的行号")
				}
			}
		})
	})
	dialog.Footer(func() {
		f.Row().Gap(8).Justify(flex.End).Children(func() {
			f.Button(goLineClose, m.t("取消")).Outline()
			if f.Button(goLineConfirm, m.t("转到")).Clicked() {
				if m.goToLine(m.goLineField.Text()) {
					m.goLineOpen = false
				} else {
					m.status = m.t("无效的行号")
				}
			}
		})
	})
	if dialog.Dismissed() {
		m.goLineOpen = false
	}
	_ = theme
}

func (m *Model) paintWelcome(f *flex.Frame, theme vivaui.Theme) {
	f.Column().Grow(1).Center().Children(func() {
		f.Column().Gap(8).Children(func() {
			paintEditorHint(f, m.t("打开 JAR 或 Class 文件"), "Ctrl+O", linkColor(m.Theme), theme)
			paintEditorHint(f, m.t("筛选项目"), "Ctrl+Shift+F", linkColor(m.Theme), theme)
			paintEditorHint(f, m.t("在源码中查找"), "Ctrl+F", linkColor(m.Theme), theme)
			paintEditorHint(f, m.t("转到行"), "Ctrl+G", linkColor(m.Theme), theme)
			f.Text(m.t("将文件拖放到此处即可打开")).FontSize(13).Muted().TextCenter().MarginSides(6, 0, 0, 0)
		})
	})
}

func paintEditorHint(f *flex.Frame, label, keys string, keyColor vivaui.Color, theme vivaui.Theme) {
	f.Row().Gap(16).Items(flex.Center).Children(func() {
		f.Text(label).Width(160).FontSize(13).TextEnd().Color(theme.MutedForeground)
		f.Text(keys).Width(112).FontSize(13).Color(keyColor)
	})
}
