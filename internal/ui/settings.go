package ui

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
)

type settingsPage struct {
	title       string
	parent      int
	group       bool
	description string
	labels      []string
}

var (
	settingsPages = []settingsPage{
		{title: "外观与行为", parent: -1, group: true, description: "自定义 JarLens 的主题、界面字体和启动行为。"},
		{title: "外观", parent: 0, labels: []string{"主题", "界面字体", "界面语言"}},
		{title: "编辑器", parent: -1, group: true, description: "调整代码字体、字号和编辑器显示。"},
		{title: "字体", parent: 2, labels: []string{"代码字体", "字号", "字体大小"}},
		{title: "常规", parent: 2, labels: []string{"显示行号", "高亮当前行"}},
		{title: "系统设置", parent: 0, labels: []string{"Java 可执行文件", "可执行文件", "启动时恢复最近项目"}},
	}
	settingsPageChildren = [][]int{{1, 5}, nil, {3, 4}, nil, nil, nil}
	settingsPageRoots    = []int{0, 2}
)

func (m *Model) openSettings() {
	m.mainMenuOpen = false
	m.settingsDraft = m.settings
	m.settingsJavaText = m.settings.JavaPath
	m.settingsJava.SetText(m.settingsJavaText)
	m.settingsFontSizeText = strconv.Itoa(m.settings.FontSize)
	m.settingsFontSize.SetText(m.settingsFontSizeText)
	page := m.settingsCategory
	if page <= 0 || page >= len(settingsPages) {
		page = 1
	}
	m.settingsCategory = page
	m.settingsHistory = []int{page}
	m.settingsHistoryIndex = 0
	m.settingsAppearanceExpanded = true
	m.settingsEditorExpanded = true
	m.settingsThemeSelect.Open = false
	m.settingsError = ""
	m.settingsQuery = ""
	m.settingsSearch.SetText("")
	m.settingsOpen = true
}

func (m *Model) closeSettings() {
	m.settingsOpen = false
	m.settingsThemeSelect.Open = false
	m.settingsError = ""
}

func (m *Model) syncSettingsFields() {
	if query := m.settingsSearch.Text(); query != m.settingsQuery {
		m.settingsQuery = query
		m.filterSettingsPages(query)
	}
	if java := m.settingsJava.Text(); java != m.settingsJavaText {
		m.settingsJavaText = java
		if strings.Contains(m.settingsError, "Java") {
			m.settingsError = ""
		}
	}
	text := m.settingsFontSize.Text()
	if text == m.settingsFontSizeText {
		return
	}
	m.settingsFontSizeText = text
	value, err := strconv.Atoi(strings.TrimSpace(text))
	if err == nil && value >= 10 && value <= 24 {
		m.settingsDraft.FontSize = value
		if strings.Contains(m.settingsError, "字号") {
			m.settingsError = ""
		}
	} else {
		m.settingsError = "字号应为 10 到 24 的整数"
	}
}

func (m *Model) applySettings(resetEditors bool) {
	if m.settings.Theme == "light" {
		m.Theme = AppearanceLight
	} else {
		m.Theme = AppearanceDark
	}
	switch {
	case m.settings.UIFont == "system":
		app.SetInterfaceFontFamily("")
	case strings.HasPrefix(m.settings.UIFont, "family:"):
		app.SetInterfaceFontFamily(strings.TrimPrefix(m.settings.UIFont, "family:"))
	default:
		app.SetInterfaceFontFamily(interfaceFamily)
	}
	if m.settings.CodeFont == "system" {
		m.codeFace = vivaui.MonoSystem
	} else if strings.HasPrefix(m.settings.CodeFont, "family:") {
		m.codeFace = vivaui.MonoFamily(strings.TrimPrefix(m.settings.CodeFont, "family:"))
	} else {
		m.codeFace = vivaui.MonoFamily(jetBrainsFamily)
	}
	if engine, ok := m.engine.(interface{ SetJava(string) }); ok {
		engine.SetJava(m.settings.JavaPath)
	}
	if resetEditors {
		m.walkPanes(func(pane *editorPane) {
			for _, doc := range pane.documents {
				doc.code = components.CodeEditor{}
			}
			pane.code = components.CodeEditor{}
		})
		m.code = components.CodeEditor{}
	}
}

func (m *Model) saveSettings(closeDialog bool) {
	fontSize, err := strconv.Atoi(strings.TrimSpace(m.settingsFontSize.Text()))
	if err != nil || fontSize < 10 || fontSize > 24 {
		m.settingsError = "字号应为 10 到 24 的整数"
		return
	}
	if !m.settingsDirty() {
		if closeDialog {
			m.closeSettings()
		}
		return
	}
	m.settingsDraft.FontSize = fontSize
	m.settingsDraft.JavaPath = strings.TrimSpace(m.settingsJava.Text())
	if m.settingsDraft.JavaPath != "" {
		info, err := os.Stat(m.settingsDraft.JavaPath)
		if err != nil || !info.Mode().IsRegular() {
			m.settingsError = "Java 可执行文件路径无效"
			return
		}
	}
	if m.settingsDraft.Theme != "light" {
		m.settingsDraft.Theme = "dark"
	}
	resetEditors := m.settingsDraft.FontSize != m.settings.FontSize || m.settingsDraft.CodeFont != m.settings.CodeFont
	m.settings = m.settingsDraft
	if m.SettingsStore != nil {
		m.SettingsStore.Save(m.settings)
	}
	m.applySettings(resetEditors)
	m.settingsJavaText = m.settings.JavaPath
	m.settingsJava.SetText(m.settingsJavaText)
	m.settingsFontSizeText = strconv.Itoa(m.settings.FontSize)
	m.settingsFontSize.SetText(m.settingsFontSizeText)
	m.settingsError = ""
	m.status = "设置已保存"
	if closeDialog {
		m.closeSettings()
	}
}

func (m *Model) settingsDirty() bool {
	for page := range settingsPages {
		if !settingsPages[page].group && m.settingsPageDirty(page) {
			return true
		}
	}
	return false
}

func (m *Model) settingsPageDirty(page int) bool {
	switch page {
	case 1:
		return m.settingsDraft.Theme != m.settings.Theme || m.settingsDraft.UIFont != m.settings.UIFont
	case 3:
		fontSize, err := strconv.Atoi(strings.TrimSpace(m.settingsFontSize.Text()))
		invalidFontSize := err != nil || fontSize < 10 || fontSize > 24
		return invalidFontSize || m.settingsDraft.CodeFont != m.settings.CodeFont || m.settingsFontSizeValue() != m.settings.FontSize
	case 4:
		return m.settingsDraft.ShowLineNumbers != m.settings.ShowLineNumbers || m.settingsDraft.HighlightCurrentLine != m.settings.HighlightCurrentLine
	case 5:
		return strings.TrimSpace(m.settingsJava.Text()) != m.settings.JavaPath || m.settingsDraft.RestoreLastProject != m.settings.RestoreLastProject
	default:
		return false
	}
}

func (m *Model) settingsNavigationDirty(page int) bool {
	if m.settingsPageDirty(page) {
		return true
	}
	for _, child := range settingsPageChildren[page] {
		if m.settingsNavigationDirty(child) {
			return true
		}
	}
	return false
}

func (m *Model) settingsFontSizeValue() int {
	value, err := strconv.Atoi(strings.TrimSpace(m.settingsFontSize.Text()))
	if err != nil || value < 10 || value > 24 {
		return m.settingsDraft.FontSize
	}
	return value
}

func (m *Model) resetSettingsPage(page int) {
	switch page {
	case 1:
		m.settingsDraft.Theme = m.settings.Theme
		m.settingsDraft.UIFont = m.settings.UIFont
	case 3:
		m.settingsDraft.CodeFont = m.settings.CodeFont
		m.settingsDraft.FontSize = m.settings.FontSize
		m.settingsFontSizeText = strconv.Itoa(m.settings.FontSize)
		m.settingsFontSize.SetText(m.settingsFontSizeText)
	case 4:
		m.settingsDraft.ShowLineNumbers = m.settings.ShowLineNumbers
		m.settingsDraft.HighlightCurrentLine = m.settings.HighlightCurrentLine
	case 5:
		m.settingsDraft.RestoreLastProject = m.settings.RestoreLastProject
		m.settingsJavaText = m.settings.JavaPath
		m.settingsJava.SetText(m.settingsJavaText)
	}
	m.settingsError = ""
}

func (m *Model) selectSettingsPage(page int) {
	if page < 0 || page >= len(settingsPages) {
		return
	}
	if m.settingsHistoryIndex >= 0 && m.settingsHistoryIndex < len(m.settingsHistory) && m.settingsHistory[m.settingsHistoryIndex] == page {
		m.settingsCategory = page
		return
	}
	m.settingsCategory = page
	if m.settingsHistoryIndex < 0 {
		m.settingsHistory = nil
		m.settingsHistoryIndex = -1
	}
	m.settingsHistory = append(m.settingsHistory[:m.settingsHistoryIndex+1], page)
	m.settingsHistoryIndex = len(m.settingsHistory) - 1
	m.view.ScrollState(settingsContentScroll).ScrollTo(0)
}

func (m *Model) goSettingsHistory(index int) {
	if index < 0 || index >= len(m.settingsHistory) {
		return
	}
	m.settingsHistoryIndex = index
	m.settingsCategory = m.settingsHistory[index]
	m.view.ScrollState(settingsContentScroll).ScrollTo(0)
}

func (m *Model) filterSettingsPages(query string) {
	visible := m.settingsVisiblePages()
	if strings.TrimSpace(query) == "" || m.settingsPageMatches(m.settingsCategory, query) {
		return
	}
	for _, page := range visible {
		if !settingsPages[page].group {
			m.selectSettingsPage(page)
			return
		}
	}
}

func (m *Model) moveSettingsSelection(delta int) {
	visible := m.settingsVisiblePages()
	if len(visible) == 0 {
		return
	}
	index := 0
	for position, page := range visible {
		if page == m.settingsCategory {
			index = position
			break
		}
	}
	index = max(0, min(len(visible)-1, index+delta))
	m.selectSettingsPage(visible[index])
}

func (m *Model) settingsVisiblePages() []int {
	query := m.settingsSearch.Text()
	visible := make([]int, 0, len(settingsPages))
	for _, root := range settingsPageRoots {
		children := make([]int, 0, len(settingsPageChildren[root]))
		for _, child := range settingsPageChildren[root] {
			if strings.TrimSpace(query) == "" || m.settingsPageMatches(child, query) {
				children = append(children, child)
			}
		}
		if len(children) == 0 {
			continue
		}
		visible = append(visible, root)
		expanded := m.settingsAppearanceExpanded
		if root == 2 {
			expanded = m.settingsEditorExpanded
		}
		if expanded || strings.TrimSpace(query) != "" {
			visible = append(visible, children...)
		}
	}
	return visible
}

func (m *Model) settingsPageMatches(page int, query string) bool {
	if page < 0 || page >= len(settingsPages) || settingsPages[page].group {
		return false
	}
	info := settingsPages[page]
	texts := []string{info.title}
	for parent := info.parent; parent >= 0; parent = settingsPages[parent].parent {
		texts = append(texts, settingsPages[parent].title)
	}
	texts = append(texts, info.labels...)
	terms := settingsQueryTerms(query)
	for _, term := range terms {
		matched := false
		for _, text := range texts {
			if strings.Contains(strings.ToLower(text), term) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return len(terms) > 0
}

func settingsQueryTerms(query string) []string {
	return strings.Fields(strings.ToLower(strings.TrimSpace(query)))
}

func settingsLabelMatches(label, query string) bool {
	lower := strings.ToLower(label)
	for _, term := range settingsQueryTerms(query) {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func (m *Model) paintSettingsDialog(f *flex.Frame, theme vivaui.Theme) {
	m.syncSettingsFields()
	theme.Border = contextMenuTheme(theme, m.Theme).Border
	dialog := f.Dialog(flex.DialogSpec{
		Title: "设置", Width: 980, Height: 680,
		Scope: settingsDialog, DismissID: settingsDismiss, CloseID: settingsClose, InitialFocusID: settingsSearch,
	})
	dialog.Body(func() {
		f.Row().Grow(1).Children(func() {
			m.paintSettingsSidebar(f, theme)
			f.Divider()
			f.Column().Grow(1).Children(func() {
				m.paintSettingsHeader(f, theme)
				m.paintSettingsBody(f, theme)
			})
		})
	})
	dialog.Footer(func() {
		if m.settingsError != "" {
			f.Text(m.settingsError).FontSize(12).Color(theme.Danger).Ellipsis()
		}
		f.Spacer()
		if runtime.GOOS == "darwin" {
			m.paintSettingsCancelButton(f)
			m.paintSettingsApplyButton(f)
			m.paintSettingsConfirmButton(f)
		} else {
			m.paintSettingsConfirmButton(f)
			m.paintSettingsCancelButton(f)
			m.paintSettingsApplyButton(f)
		}
	})
	if dialog.Dismissed() {
		m.closeSettings()
	}
}

func (m *Model) paintSettingsConfirmButton(f *flex.Frame) {
	if f.Button(settingsConfirm, "确定").Clicked() {
		m.saveSettings(true)
	}
}

func (m *Model) paintSettingsCancelButton(f *flex.Frame) {
	if f.Button(settingsCancel, "取消").Secondary().Clicked() {
		m.closeSettings()
	}
}

func (m *Model) paintSettingsApplyButton(f *flex.Frame) {
	if f.Button(settingsApply, "应用").Secondary().Disabled(!m.settingsDirty()).Clicked() {
		m.saveSettings(false)
	}
}

func (m *Model) paintSettingsSidebar(f *flex.Frame, theme vivaui.Theme) {
	f.Column().Width(240).Children(func() {
		f.Row().Pad(10).Children(func() {
			f.Field(&m.settingsSearch, settingsSearch, "搜索设置", "搜索设置项").FillWidth()
		})
		f.ScrollArea(settingsNavScroll).Grow(1).PadSides(0, 0, 6, 0).Children(func() {
			visible := m.settingsVisiblePages()
			if len(visible) == 0 {
				f.Text("没有匹配的设置项").FontSize(13).Muted().PadSides(8, 12, 0, 12)
				return
			}
			for _, page := range visible {
				m.paintSettingsNavigationRow(f, theme, page)
			}
		})
	})
}

func (m *Model) paintSettingsNavigationRow(f *flex.Frame, theme vivaui.Theme, page int) {
	info := settingsPages[page]
	selected := page == m.settingsCategory
	depth := 0
	if info.parent >= 0 {
		depth = 1
	}
	row := f.Row().Height(24).PadSides(0, 8, 0, float32(4+depth*16)).Items(flex.Center).
		Hover(flex.Look{}.Background(theme.Hover)).
		HitWith(settingsNav.At(page), info.title, vivaui.HitOptions{Role: vivaui.RoleTab, Selected: selected, Cursor: vivaui.PointerCursorHand})
	if selected {
		row = row.Background(theme.Accent)
	}
	row.Children(func() {
		if info.group {
			expanded := m.settingsAppearanceExpanded
			if page == 2 {
				expanded = m.settingsEditorExpanded
			}
			glyph := chevronRightGlyph
			if expanded || strings.TrimSpace(m.settingsSearch.Text()) != "" {
				glyph = chevronDownGlyph
			}
			button := f.Row().Size(16, 24).Center().HitWith(settingsNavDisclosures.At(page), "展开或折叠 "+info.title, vivaui.HitOptions{
				Role: vivaui.RoleButton, Expanded: expanded, Expandable: true, Cursor: vivaui.PointerCursorHand,
			})
			button.Children(func() {
				f.Icon(glyph).Size(14, 14).Color(theme.MutedForeground)
			})
			if button.Clicked() {
				if page == 0 {
					m.settingsAppearanceExpanded = !m.settingsAppearanceExpanded
				} else {
					m.settingsEditorExpanded = !m.settingsEditorExpanded
				}
			}
		} else {
			f.Row().Size(16, 24)
		}
		color := theme.AccentForeground
		if m.settingsNavigationDirty(page) && !selected {
			color = theme.Primary
		}
		settingsSearchLabel(f, info.title, m.settingsSearch.Text(), color, 13, info.group)
	})
	if row.Clicked() {
		m.selectSettingsPage(page)
	}
}

func (m *Model) paintSettingsHeader(f *flex.Frame, theme vivaui.Theme) {
	info := settingsPages[m.settingsCategory]
	f.Row().Height(40).PadSides(0, 12, 0, 20).Gap(12).Items(flex.Center).Children(func() {
		f.Row().Grow(1).Gap(4).Items(flex.Center).Children(func() {
			if info.parent >= 0 {
				f.Text(settingsPages[info.parent].title).FontSize(13).Weight(vivaui.WeightSemibold)
				f.Icon(chevronRightGlyph).Size(12, 12).Color(theme.MutedForeground)
			}
			f.Text(info.title).FontSize(13).Weight(vivaui.WeightSemibold).Ellipsis()
		})
		if !info.group && m.settingsPageDirty(m.settingsCategory) {
			if f.Button(settingsReset, "重置").Link().Small().Clicked() {
				m.resetSettingsPage(m.settingsCategory)
			}
		}
		f.Row().Gap(2).Items(flex.Center).Children(func() {
			if smallIconButton(f, settingsBack, "后退", settingsArrowLeftGlyph, theme, m.settingsHistoryIndex > 0) {
				m.goSettingsHistory(m.settingsHistoryIndex - 1)
			}
			if smallIconButton(f, settingsForward, "前进", settingsArrowRightGlyph, theme, m.settingsHistoryIndex >= 0 && m.settingsHistoryIndex < len(m.settingsHistory)-1) {
				m.goSettingsHistory(m.settingsHistoryIndex + 1)
			}
		})
	})
}

func (m *Model) paintSettingsBody(f *flex.Frame, theme vivaui.Theme) {
	f.ScrollArea(settingsContentScroll).Grow(1).PadSides(4, 20, 20, 20).Children(func() {
		f.Column().FillWidth().MaxWidth(680).Children(func() {
			switch m.settingsCategory {
			case 0, 2:
				m.paintSettingsGroupPage(f, m.settingsCategory)
			case 1:
				m.paintAppearanceSettings(f, theme)
			case 3:
				m.paintEditorFontSettings(f, theme)
			case 4:
				m.paintEditorGeneralSettings(f, theme)
			case 5:
				m.paintSystemSettings(f, theme)
			}
		})
	})
}

func (m *Model) paintSettingsGroupPage(f *flex.Frame, page int) {
	f.Column().PadSides(8, 0, 0, 0).Children(func() {
		f.Paragraph(settingsPages[page].description, 3).FontSize(13).MarginSides(0, 0, 14, 0)
		f.Column().PadSides(0, 0, 0, 32).Gap(2).Children(func() {
			for _, child := range settingsPageChildren[page] {
				if f.Button(settingsGroupLinks.At(child), settingsPages[child].title).Link().Height(20).Clicked() {
					m.selectSettingsPage(child)
				}
			}
		})
	})
}

func (m *Model) paintAppearanceSettings(f *flex.Frame, theme vivaui.Theme) {
	settingsGroup(f, theme, "", func() {
		m.settingsRow(f, theme, "主题", func() {
			selected := 0
			if m.settingsDraft.Theme == "light" {
				selected = 1
			}
			selector := f.Select(settingsTheme, &m.settingsThemeSelect, []string{"Islands Dark", "Islands Light"}, selected).Width(240).Label("主题")
			if m.settingsThemeSelect.Open {
				// Close the select's presence before the enclosing dialog closes its own.
				selector = selector.DrawOver(func(d *flex.DrawContext) { d.Canvas.EndPresence() })
			}
			if index, ok := selector.Picked(); ok {
				m.settingsDraft.Theme = "dark"
				if index == 1 {
					m.settingsDraft.Theme = "light"
				}
			}
		})
		m.settingsRow(f, theme, "界面字体", func() {
			if f.Button(settingsUIFontChoose, m.fontDisplayFamily(0)).Secondary().Width(240).Clicked() {
				f.Canvas().AfterPaint(func() { m.openFontPicker(0) })
			}
		})
	})
	settingsGroup(f, theme, "语言", func() {
		m.settingsRow(f, theme, "界面语言", func() {
			f.Text("简体中文").FontSize(13).Muted()
		})
	})
}

func (m *Model) paintEditorFontSettings(f *flex.Frame, theme vivaui.Theme) {
	settingsGroup(f, theme, "", func() {
		m.settingsRow(f, theme, "代码字体", func() {
			if f.Button(settingsCodeFontChoose, m.fontDisplayFamily(1)).Secondary().Width(240).Clicked() {
				f.Canvas().AfterPaint(func() { m.openFontPicker(1) })
			}
		})
		m.settingsRow(f, theme, "字号", func() {
			f.Row().Gap(6).Items(flex.Center).Children(func() {
				f.Field(&m.settingsFontSize, settingsFontSizeInput, "字号", "13").Width(64)
				f.Text("px").FontSize(13).Muted()
			})
		})
	})
}

func (m *Model) paintEditorGeneralSettings(f *flex.Frame, theme vivaui.Theme) {
	settingsGroup(f, theme, "显示", func() {
		m.settingsToggleRow(f, theme, settingsLineNumbers, "显示行号", m.settingsDraft.ShowLineNumbers, func() {
			m.settingsDraft.ShowLineNumbers = !m.settingsDraft.ShowLineNumbers
		})
		m.settingsToggleRow(f, theme, settingsCurrentLine, "高亮当前行", m.settingsDraft.HighlightCurrentLine, func() {
			m.settingsDraft.HighlightCurrentLine = !m.settingsDraft.HighlightCurrentLine
		})
	})
}

func (m *Model) paintSystemSettings(f *flex.Frame, theme vivaui.Theme) {
	settingsGroup(f, theme, "Java 运行环境", func() {
		m.settingsRow(f, theme, "可执行文件", func() {
			f.Field(&m.settingsJava, settingsJavaPath, "Java 可执行文件", "自动检测 PATH 或内置 JRE").Width(360)
		})
	})
	settingsGroup(f, theme, "启动", func() {
		m.settingsToggleRow(f, theme, settingsRestore, "启动时恢复最近项目", m.settingsDraft.RestoreLastProject, func() {
			m.settingsDraft.RestoreLastProject = !m.settingsDraft.RestoreLastProject
		})
	})
}

func settingsGroup(f *flex.Frame, theme vivaui.Theme, title string, children func()) {
	f.Column().MarginSides(0, 0, 22, 0).Children(func() {
		if title != "" {
			f.Row().Height(22).MarginSides(0, 0, 6, 0).Gap(8).Items(flex.Center).Children(func() {
				f.Text(title).FontSize(13).Weight(vivaui.WeightSemibold)
				f.Row().Grow(1).Height(1).Background(theme.Border)
			})
		}
		f.Column().PadSides(0, 0, 0, 18).Children(children)
	})
}

func (m *Model) settingsRow(f *flex.Frame, theme vivaui.Theme, label string, children func()) {
	matched := settingsLabelMatches(label, m.settingsSearch.Text())
	row := f.Column().PadSides(2, 6, 2, 6).MarginSides(0, -6, 0, -6).Radius(4)
	if matched {
		row = row.Background(settingsRowMatchColor)
	}
	row.Children(func() {
		f.Row().MinHeight(28).Gap(8).Items(flex.Center).Children(func() {
			settingsSearchLabel(f, label+"：", m.settingsSearch.Text(), theme.Foreground, 13, false)
			f.Row().Grow(1).Items(flex.Center).Children(children)
		})
		if description := m.settingsDescription(label); description != "" {
			f.Text(description).FontSize(12).Muted().PadSides(0, 0, 2, 0)
		}
	})
}

func (m *Model) settingsToggleRow(f *flex.Frame, theme vivaui.Theme, id vivaui.ID, label string, value bool, toggle func()) {
	matched := settingsLabelMatches(label, m.settingsSearch.Text())
	row := f.Column().PadSides(2, 6, 2, 6).MarginSides(0, -6, 0, -6).Radius(4)
	if matched {
		row = row.Background(settingsRowMatchColor)
	}
	row.Children(func() {
		f.Row().MinHeight(26).Gap(6).Items(flex.Center).Children(func() {
			if f.Checkbox(id, value).Label(label).Clicked() {
				toggle()
			}
			settingsSearchLabel(f, label, m.settingsSearch.Text(), theme.Foreground, 13, false)
		})
		if description := m.settingsDescription(label); description != "" {
			f.Text(description).FontSize(12).Muted().PadSides(0, 0, 2, 26)
		}
	})
}

var settingsRowMatchColor = vivaui.RGBA(242, 197, 92, 0.12)

func settingsSearchLabel(f *flex.Frame, text, query string, color vivaui.Color, size float32, bold bool) {
	terms := settingsQueryTerms(query)
	if len(terms) == 0 {
		settingsLabelText(f, text, color, size, bold)
		return
	}
	f.Row().Gap(0).Items(flex.Center).Children(func() {
		remaining := text
		for remaining != "" {
			start, length := settingsFirstMatch(remaining, terms)
			if start < 0 {
				settingsLabelText(f, remaining, color, size, bold)
				return
			}
			if start > 0 {
				settingsLabelText(f, remaining[:start], color, size, bold)
			}
			settingsLabelText(f, remaining[start:start+length], color, size, bold).Background(settingsRowMatchColor).Radius(2)
			remaining = remaining[start+length:]
		}
	})
}

func settingsLabelText(f *flex.Frame, text string, color vivaui.Color, size float32, bold bool) flex.Box {
	box := f.Text(text).FontSize(size).Color(color)
	if bold {
		box = box.Weight(vivaui.WeightSemibold)
	}
	return box
}

func settingsFirstMatch(text string, terms []string) (int, int) {
	lower := strings.ToLower(text)
	best, length := -1, 0
	for _, term := range terms {
		if term == "" {
			continue
		}
		if index := strings.Index(lower, term); index >= 0 && (best < 0 || index < best || index == best && len(term) > length) {
			best, length = index, len(term)
		}
	}
	return best, length
}

func (m *Model) settingsDescription(label string) string {
	switch label {
	case "主题":
		return "选择 JarLens 的外观颜色。"
	case "界面字体":
		return "选择已安装的界面字体。"
	case "界面语言":
		return "当前版本使用简体中文。"
	case "代码字体":
		return "选择已安装的等宽字体。"
	case "字号":
		return "设置编辑器的代码字号，范围为 10 到 24 px。"
	case "显示行号":
		return "在编辑器左侧显示行号。"
	case "高亮当前行":
		return "高亮显示光标所在的代码行。"
	case "可执行文件":
		return "留空时自动检测内置 JRE、JARLENS_JAVA 或 PATH。"
	case "启动时恢复最近项目":
		return "下次启动 JarLens 时恢复上次打开的项目。"
	default:
		return ""
	}
}

func (m *Model) stepSettingsFontSize(delta int) {
	value := m.settingsFontSizeValue() + delta
	value = max(10, min(24, value))
	m.settingsDraft.FontSize = value
	m.settingsFontSizeText = strconv.Itoa(value)
	m.settingsFontSize.SetText(m.settingsFontSizeText)
	m.settingsError = ""
}

var settingsArrowLeftGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(19, 12), vivaui.LineTo(5, 12),
		vivaui.MoveTo(11, 18), vivaui.LineTo(5, 12), vivaui.LineTo(11, 6),
	},
}

var settingsArrowRightGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(5, 12), vivaui.LineTo(19, 12),
		vivaui.MoveTo(13, 6), vivaui.LineTo(19, 12), vivaui.LineTo(13, 18),
	},
}
