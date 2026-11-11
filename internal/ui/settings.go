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
	jarlensupdate "github.com/xddcode/jarlens/internal/update"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
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
		{title: "常规", parent: 2, labels: []string{"Tab 宽度", "缩进", "显示缩进参考线", "显示行号", "高亮当前行"}},
		{title: "系统设置", parent: 0, labels: []string{"Java 可执行文件", "可执行文件", "反编译引擎", "Vineflower", "CFR", "Procyon", "启动时", "恢复最近项目", "回到 Home", "关闭时最小化到托盘", "开机自动启动"}},
		{title: "更新管理", parent: -1, labels: []string{"更新下载源", "默认下载源", "CNB", "自动检查更新", "检查更新", "当前版本", "更新日志", "加载更多"}},
	}
	settingsPageChildren = [][]int{{1, 5}, nil, {3, 4}, nil, nil, nil, nil}
	settingsPageRoots    = []int{0, 2, 6}
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
	m.settingsLanguageSelect.Open = false
	m.settingsUpdateSourceSelect.Open = false
	m.settingsEngineSelect.Open = false
	m.settingsError = ""
	m.settingsQuery = ""
	m.settingsSearch.SetText("")
	m.settingsOpen = true
	if page == 6 {
		m.updateReleasesLoaded = false
	}
}

func (m *Model) closeSettings() {
	m.settingsOpen = false
	m.settingsThemeSelect.Open = false
	m.settingsLanguageSelect.Open = false
	m.settingsUpdateSourceSelect.Open = false
	m.settingsEngineSelect.Open = false
	m.settingsError = ""
	m.cancelUpdateReleases()
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

func (m *Model) applyThemePreference(preference string) {
	m.Theme = AppearanceDark
	nativeAppearance := app.AppearanceDark
	switch preference {
	case "light":
		m.Theme = AppearanceLight
		nativeAppearance = app.AppearanceLight
	case "system":
		nativeAppearance = app.AppearanceUnknown
		if m.stopSystemAppearance != nil && app.SystemAppearance() == app.AppearanceLight {
			m.Theme = AppearanceLight
		}
	}
	m.nativeAppearance = nativeAppearance
	m.nativeAppearanceDirty = true
}

func (m *Model) SyncNativeAppearance() {
	if m.stopSystemAppearance == nil && m.invalidator != nil {
		m.stopSystemAppearance = app.WatchSystemAppearance(m.invalidator.Invalidate)
	}
	if m.settings.Theme == "system" {
		m.applyThemePreference("system")
	}
	if !m.nativeAppearanceDirty {
		return
	}
	app.SetNativeAppearance(m.nativeAppearance)
	m.nativeAppearanceDirty = false
}

func (m *Model) applySettings(resetEditors bool) {
	m.applyThemePreference(m.settings.Theme)
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
	m.selectEngine(m.settings.Engine)
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
	switch m.settingsDraft.Theme {
	case "light", "dark", "system":
	default:
		m.settingsDraft.Theme = "dark"
	}
	if m.settingsDraft.TabWidth != 2 && m.settingsDraft.TabWidth != 4 && m.settingsDraft.TabWidth != 6 {
		m.settingsDraft.TabWidth = 4
	}
	m.settingsDraft.Language = normalizeLanguage(m.settingsDraft.Language)
	m.settingsDraft.UpdateSource = string(jarlensupdate.ParseSource(m.settingsDraft.UpdateSource))
	if decompilerEngineIndex(m.settingsDraft.Engine) < 0 {
		m.settingsDraft.Engine = "vineflower"
	}
	if m.settingsDraft.LaunchAtLogin != m.settings.LaunchAtLogin {
		setLaunchAtLogin := m.SetLaunchAtLogin
		if setLaunchAtLogin == nil {
			setLaunchAtLogin = jarlensworkspace.SetLaunchAtLogin
		}
		if err := setLaunchAtLogin(m.settingsDraft.LaunchAtLogin); err != nil {
			m.settingsError = "无法设置开机自动启动"
			m.diagnostic = err.Error()
			return
		}
	}
	resetEditors := m.settingsDraft.FontSize != m.settings.FontSize || m.settingsDraft.CodeFont != m.settings.CodeFont
	previousTabWidth := m.settings.TabWidth
	m.settings = m.settingsDraft
	if m.SettingsStore != nil {
		m.SettingsStore.Save(m.settings)
	}
	m.applySettings(resetEditors)
	m.SyncNativeAppearance()
	m.reindentDocuments(previousTabWidth, m.settings.TabWidth)
	m.settingsJavaText = m.settings.JavaPath
	m.settingsJava.SetText(m.settingsJavaText)
	m.settingsFontSizeText = strconv.Itoa(m.settings.FontSize)
	m.settingsFontSize.SetText(m.settingsFontSizeText)
	m.settingsError = ""
	m.status = m.t("设置已保存")
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
		return m.settingsDraft.Theme != m.settings.Theme || m.settingsDraft.Language != m.settings.Language || m.settingsDraft.UIFont != m.settings.UIFont
	case 3:
		fontSize, err := strconv.Atoi(strings.TrimSpace(m.settingsFontSize.Text()))
		invalidFontSize := err != nil || fontSize < 10 || fontSize > 24
		return invalidFontSize || m.settingsDraft.CodeFont != m.settings.CodeFont || m.settingsFontSizeValue() != m.settings.FontSize
	case 4:
		return m.settingsDraft.TabWidth != m.settings.TabWidth || m.settingsDraft.ShowIndentGuides != m.settings.ShowIndentGuides || m.settingsDraft.ShowLineNumbers != m.settings.ShowLineNumbers || m.settingsDraft.HighlightCurrentLine != m.settings.HighlightCurrentLine
	case 5:
		return strings.TrimSpace(m.settingsJava.Text()) != m.settings.JavaPath || m.settingsDraft.RestoreLastProject != m.settings.RestoreLastProject || m.settingsDraft.CloseToTray != m.settings.CloseToTray || m.settingsDraft.LaunchAtLogin != m.settings.LaunchAtLogin || m.settingsDraft.Engine != m.settings.Engine
	case 6:
		return m.settingsDraft.CheckUpdates != m.settings.CheckUpdates || m.selectedUpdateSource() != jarlensupdate.ParseSource(m.settings.UpdateSource)
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
		m.settingsDraft.Language = m.settings.Language
		m.settingsDraft.UIFont = m.settings.UIFont
	case 3:
		m.settingsDraft.CodeFont = m.settings.CodeFont
		m.settingsDraft.FontSize = m.settings.FontSize
		m.settingsFontSizeText = strconv.Itoa(m.settings.FontSize)
		m.settingsFontSize.SetText(m.settingsFontSizeText)
	case 4:
		m.settingsDraft.TabWidth = m.settings.TabWidth
		m.settingsDraft.ShowIndentGuides = m.settings.ShowIndentGuides
		m.settingsDraft.ShowLineNumbers = m.settings.ShowLineNumbers
		m.settingsDraft.HighlightCurrentLine = m.settings.HighlightCurrentLine
	case 5:
		m.settingsDraft.RestoreLastProject = m.settings.RestoreLastProject
		m.settingsDraft.CloseToTray = m.settings.CloseToTray
		m.settingsDraft.LaunchAtLogin = m.settings.LaunchAtLogin
		m.settingsDraft.Engine = m.settings.Engine
		m.settingsJavaText = m.settings.JavaPath
		m.settingsJava.SetText(m.settingsJavaText)
	case 6:
		m.settingsDraft.CheckUpdates = m.settings.CheckUpdates
		m.settingsDraft.UpdateSource = m.settings.UpdateSource
		m.updateReleasesLoaded = false
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
	if page == 6 {
		m.updateReleasesLoaded = false
	}
	m.view.ScrollState(settingsContentScroll).ScrollTo(0)
}

func (m *Model) goSettingsHistory(index int) {
	if index < 0 || index >= len(m.settingsHistory) {
		return
	}
	m.settingsHistoryIndex = index
	m.settingsCategory = m.settingsHistory[index]
	if m.settingsCategory == 6 {
		m.updateReleasesLoaded = false
	}
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
			if !settingsPages[root].group && (strings.TrimSpace(query) == "" || m.settingsPageMatches(root, query)) {
				visible = append(visible, root)
			}
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
	texts := []string{info.title, m.t(info.title)}
	for parent := info.parent; parent >= 0; parent = settingsPages[parent].parent {
		texts = append(texts, settingsPages[parent].title, m.t(settingsPages[parent].title))
	}
	for _, label := range info.labels {
		texts = append(texts, label, m.t(label))
	}
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
		Title: m.t("设置"), Width: 980, Height: 680,
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
			f.Text(m.t(m.settingsError)).FontSize(12).Color(theme.Danger).Ellipsis()
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
	if f.Button(settingsConfirm, m.t("确定")).Clicked() {
		m.saveSettings(true)
	}
}

func (m *Model) paintSettingsCancelButton(f *flex.Frame) {
	if f.Button(settingsCancel, m.t("取消")).Secondary().Clicked() {
		m.closeSettings()
	}
}

func (m *Model) paintSettingsApplyButton(f *flex.Frame) {
	if f.Button(settingsApply, m.t("应用")).Secondary().Disabled(!m.settingsDirty()).Clicked() {
		m.saveSettings(false)
	}
}

func (m *Model) paintSettingsSidebar(f *flex.Frame, theme vivaui.Theme) {
	f.Column().Width(240).Children(func() {
		f.Row().Pad(10).Children(func() {
			f.Field(&m.settingsSearch, settingsSearch, m.t("搜索设置"), m.t("搜索设置项")).FillWidth()
		})
		f.ScrollArea(settingsNavScroll).Grow(1).PadSides(0, 0, 6, 0).Children(func() {
			visible := m.settingsVisiblePages()
			if len(visible) == 0 {
				f.Text(m.t("没有匹配的设置项")).FontSize(13).Muted().PadSides(8, 12, 0, 12)
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
		HitWith(settingsNav.At(page), m.t(info.title), vivaui.HitOptions{Role: vivaui.RoleTab, Selected: selected, Cursor: vivaui.PointerCursorHand})
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
			button := f.Row().Size(16, 24).Center().HitWith(settingsNavDisclosures.At(page), m.tf("展开或折叠 %s", info.title), vivaui.HitOptions{
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
		settingsSearchLabel(f, m.t(info.title), m.settingsSearch.Text(), color, 13, info.group)
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
				f.Text(m.t(settingsPages[info.parent].title)).FontSize(13).Weight(vivaui.WeightSemibold)
				f.Icon(chevronRightGlyph).Size(12, 12).Color(theme.MutedForeground)
			}
			f.Text(m.t(info.title)).FontSize(13).Weight(vivaui.WeightSemibold).Ellipsis()
		})
		if !info.group && m.settingsPageDirty(m.settingsCategory) {
			if f.Button(settingsReset, m.t("重置")).Link().Small().Clicked() {
				m.resetSettingsPage(m.settingsCategory)
			}
		}
		f.Row().Gap(2).Items(flex.Center).Children(func() {
			if smallIconButton(f, settingsBack, m.t("后退"), settingsArrowLeftGlyph, theme, m.settingsHistoryIndex > 0) {
				m.goSettingsHistory(m.settingsHistoryIndex - 1)
			}
			if smallIconButton(f, settingsForward, m.t("前进"), settingsArrowRightGlyph, theme, m.settingsHistoryIndex >= 0 && m.settingsHistoryIndex < len(m.settingsHistory)-1) {
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
			case 6:
				m.paintUpdateSettings(f, theme)
			}
		})
	})
}

func (m *Model) paintSettingsGroupPage(f *flex.Frame, page int) {
	f.Column().PadSides(8, 0, 0, 0).Children(func() {
		f.Paragraph(m.t(settingsPages[page].description), 3).FontSize(13).MarginSides(0, 0, 14, 0)
		f.Column().PadSides(0, 0, 0, 32).Gap(2).Children(func() {
			for _, child := range settingsPageChildren[page] {
				if f.Button(settingsGroupLinks.At(child), m.t(settingsPages[child].title)).Link().Height(20).Clicked() {
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
			switch m.settingsDraft.Theme {
			case "light":
				selected = 1
			case "system":
				selected = 2
			}
			selector := f.Select(settingsTheme, &m.settingsThemeSelect, []string{m.t("深色"), m.t("浅色"), m.t("跟随系统")}, selected).Width(240).Label(m.t("主题"))
			if m.settingsThemeSelect.Open {
				// Close the select's presence before the enclosing dialog closes its own.
				selector = selector.DrawOver(func(d *flex.DrawContext) { d.Canvas.EndPresence() })
			}
			if index, ok := selector.Picked(); ok {
				m.settingsDraft.Theme = "dark"
				switch index {
				case 1:
					m.settingsDraft.Theme = "light"
				case 2:
					m.settingsDraft.Theme = "system"
				}
			}
		})
		m.settingsRow(f, theme, "界面字体", func() {
			if f.Button(settingsUIFontChoose, m.fontDisplayFamily(0)).Secondary().Width(240).Clicked() {
				f.Canvas().AfterPaint(func() { m.openFontPicker(0) })
			}
		})
	})
	settingsGroup(f, theme, m.t("语言"), func() {
		m.settingsRow(f, theme, "界面语言", func() {
			selector := f.Select(settingsLanguage, &m.settingsLanguageSelect, languageNames, languageIndex(m.settingsDraft.Language)).Width(240).Label(m.t("界面语言"))
			if m.settingsLanguageSelect.Open {
				selector = selector.DrawOver(func(d *flex.DrawContext) { d.Canvas.EndPresence() })
			}
			if index, ok := selector.Picked(); ok {
				m.settingsDraft.Language = languageIDs[index]
			}
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
	settingsGroup(f, theme, m.t("显示"), func() {
		m.settingsRow(f, theme, "Tab 宽度", func() {
			selected := 0
			switch m.settingsDraft.TabWidth {
			case 4:
				selected = 1
			case 6:
				selected = 2
			}
			if index, ok := f.Choice(settingsTabWidth, []string{m.t("紧凑（2）"), m.t("默认（4）"), m.t("宽松（6）")}, selected).Width(240).ChosenIndex(); ok {
				m.settingsDraft.TabWidth = 2
				switch index {
				case 1:
					m.settingsDraft.TabWidth = 4
				case 2:
					m.settingsDraft.TabWidth = 6
				}
			}
		})
		m.settingsToggleRow(f, theme, settingsIndentGuides, "显示缩进参考线", m.settingsDraft.ShowIndentGuides, func() {
			m.settingsDraft.ShowIndentGuides = !m.settingsDraft.ShowIndentGuides
		})
		m.settingsToggleRow(f, theme, settingsLineNumbers, "显示行号", m.settingsDraft.ShowLineNumbers, func() {
			m.settingsDraft.ShowLineNumbers = !m.settingsDraft.ShowLineNumbers
		})
		m.settingsToggleRow(f, theme, settingsCurrentLine, "高亮当前行", m.settingsDraft.HighlightCurrentLine, func() {
			m.settingsDraft.HighlightCurrentLine = !m.settingsDraft.HighlightCurrentLine
		})
	})
}

func (m *Model) paintSystemSettings(f *flex.Frame, theme vivaui.Theme) {
	settingsGroup(f, theme, m.t("反编译"), func() {
		m.settingsRow(f, theme, "引擎", func() {
			selected := decompilerEngineIndex(m.settingsDraft.Engine)
			selector := f.Select(settingsEngine, &m.settingsEngineSelect, decompilerEngineLabels, selected).Width(240).Label(m.t("反编译引擎"))
			if m.settingsEngineSelect.Open {
				selector = selector.DrawOver(func(d *flex.DrawContext) { d.Canvas.EndPresence() })
			}
			if index, ok := selector.Picked(); ok {
				m.settingsDraft.Engine = decompilerEngineIDs[index]
			}
		})
	})
	settingsGroup(f, theme, m.t("Java 运行环境"), func() {
		m.settingsRow(f, theme, "可执行文件", func() {
			f.Row().Width(460).Gap(8).Items(flex.Center).Children(func() {
				f.Field(&m.settingsJava, settingsJavaPath, m.t("Java 可执行文件"), m.t("自动检测 PATH 或内置 JRE")).Grow(1).MaxWidth(360)
				if f.Button(settingsJavaBrowse, m.t("浏览...")).Outline().Small().Clicked() {
					f.Canvas().AfterPaint(m.chooseSettingsJava)
				}
			})
		})
	})
	settingsGroup(f, theme, m.t("启动"), func() {
		m.settingsRow(f, theme, "启动时", func() {
			selected := 0
			if !m.settingsDraft.RestoreLastProject {
				selected = 1
			}
			if index, ok := f.Choice(settingsStartup, []string{m.t("恢复最近项目"), m.t("回到 Home")}, selected).Width(240).ChosenIndex(); ok {
				m.settingsDraft.RestoreLastProject = index == 0
			}
		})
		m.settingsToggleRow(f, theme, settingsCloseToTray, "关闭时最小化到托盘", m.settingsDraft.CloseToTray, func() {
			m.settingsDraft.CloseToTray = !m.settingsDraft.CloseToTray
		})
		m.settingsToggleRow(f, theme, settingsLaunchAtLogin, "开机自动启动", m.settingsDraft.LaunchAtLogin, func() {
			m.settingsDraft.LaunchAtLogin = !m.settingsDraft.LaunchAtLogin
		})
	})
}

func (m *Model) chooseSettingsJava() {
	paths, err := app.OpenFiles(app.OpenFileOptions{
		Title:   m.t("选择 Java 可执行文件"),
		Filters: []app.FileFilter{{Name: m.t("Java 可执行文件"), Pattern: "java;java.exe"}},
	})
	if err != nil {
		m.settingsError = m.t("无法打开文件选择对话框")
		return
	}
	if len(paths) > 0 {
		m.settingsJavaText = paths[0]
		m.settingsJava.SetText(paths[0])
		m.settingsError = ""
	}
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
	localizedLabel := m.t(label)
	matched := settingsLabelMatches(localizedLabel, m.settingsSearch.Text())
	row := f.Column().PadSides(2, 6, 2, 6).MarginSides(0, -6, 0, -6).Radius(4)
	if matched {
		row = row.Background(settingsRowMatchColor)
	}
	row.Children(func() {
		f.Row().MinHeight(28).Gap(8).Items(flex.Center).Children(func() {
			settingsSearchLabel(f, localizedLabel+m.t("："), m.settingsSearch.Text(), theme.Foreground, 13, false)
			f.Row().Grow(1).Items(flex.Center).Children(children)
		})
		if description := m.t(m.settingsDescription(label)); description != "" {
			f.Text(description).FontSize(12).Muted().PadSides(0, 0, 2, 0)
		}
	})
}

func (m *Model) settingsToggleRow(f *flex.Frame, theme vivaui.Theme, id vivaui.ID, label string, value bool, toggle func()) {
	localizedLabel := m.t(label)
	matched := settingsLabelMatches(localizedLabel, m.settingsSearch.Text())
	row := f.Column().PadSides(2, 6, 2, 6).MarginSides(0, -6, 0, -6).Radius(4)
	if matched {
		row = row.Background(settingsRowMatchColor)
	}
	row.Children(func() {
		f.Row().MinHeight(26).Gap(6).Items(flex.Center).Children(func() {
			if f.Checkbox(id, value).Label(localizedLabel).Clicked() {
				toggle()
			}
			settingsSearchLabel(f, localizedLabel, m.settingsSearch.Text(), theme.Foreground, 13, false)
		})
		if description := m.t(m.settingsDescription(label)); description != "" {
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
		return "选择深色、浅色，或跟随系统自动切换外观。"
	case "界面字体":
		return "选择已安装的界面字体。"
	case "界面语言":
		return "选择界面显示语言。"
	case "代码字体":
		return "选择已安装的等宽字体。"
	case "字号":
		return "设置编辑器的代码字号，范围为 10 到 24 px。"
	case "Tab 宽度":
		return "按紧凑 2 列、默认 4 列或宽松 6 列重排代码缩进和参考线，不修改原始归档。"
	case "显示缩进参考线":
		return "显示或隐藏代码缩进处的竖线，不改变代码缩进。"
	case "显示行号":
		return "在编辑器左侧显示行号。"
	case "高亮当前行":
		return "高亮显示光标所在的代码行。"
	case "引擎":
		return "选择用于后续反编译的引擎；已有文档保留其实际来源。"
	case "可执行文件":
		return "留空时自动检测内置 JRE、JARLENS_JAVA 或 PATH。"
	case "启动时恢复最近项目":
		return "下次启动 JarLens 时恢复上次打开的项目。"
	case "启动时":
		return "选择下次启动时恢复最近项目，或回到 Home。"
	case "关闭时最小化到托盘":
		return "关闭窗口时保留应用在托盘中运行；可从托盘菜单退出。"
	case "开机自动启动":
		return "登录系统时自动启动 JarLens。"
	case "更新下载源":
		return "默认下载源为官方源，CNB 适合国内网络环境；切换后检查更新和更新日志均使用所选来源。"
	case "自动检查更新":
		return "启动时从所选下载源检查最新正式版，不会自动安装。"
	case "检查更新":
		return "立即检查最新正式版并查看更新说明。"
	case "当前版本":
		return "当前 JarLens 的版本号。"
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
