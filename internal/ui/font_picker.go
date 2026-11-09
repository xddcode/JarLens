package ui

import (
	"sort"
	"strings"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) openFontPicker(target int) {
	m.fontPickerTarget = target
	m.fontPickerFamilies = append([]string(nil), app.FontFamilies()...)
	sort.Strings(m.fontPickerFamilies)
	m.fontPickerSearch.SetText("")
	m.fontPickerOpen = true
}

func (m *Model) fontPickerOptions() []string {
	options := []string{"Inter", "系统默认"}
	if m.fontPickerTarget == 1 {
		options = []string{"JetBrains Mono", "系统等宽"}
	}
	seen := make(map[string]bool, len(options))
	for _, option := range options {
		seen[option] = true
	}
	for _, family := range m.fontPickerFamilies {
		if !seen[family] {
			seen[family] = true
			options = append(options, family)
		}
	}
	return options
}

func (m *Model) paintFontPicker(f *flex.Frame, theme vivaui.Theme) {
	dialog := f.Dialog(flex.DialogSpec{
		Title: "选择字体", Width: 520, Height: 500, Layer: 4,
		Scope: fontPickerDialog, DismissID: fontPickerDismiss, CloseID: fontPickerClose, InitialFocusID: fontPickerSearch,
	})
	dialog.Body(func() {
		f.Column().Grow(1).Gap(12).Children(func() {
			f.Field(&m.fontPickerSearch, fontPickerSearch, "搜索字体", "输入字体名称").FillWidth()
			query := strings.ToLower(strings.TrimSpace(m.fontPickerSearch.Text()))
			f.ScrollArea(fontPickerList).Grow(1).Gap(4).Children(func() {
				found := false
				for index, family := range m.fontPickerOptions() {
					if query != "" && !strings.Contains(strings.ToLower(family), query) {
						continue
					}
					found = true
					id := settingsUIFont.At(index)
					if m.fontPickerTarget == 1 {
						id = settingsCodeFont.At(index)
					}
					selected := family == m.fontDisplayFamily(m.fontPickerTarget)
					row := f.Row().Height(32).PadX(12).Items(flex.Center).Radius(5).
						Hover(flex.Look{}.Background(theme.Hover)).
						HitWith(id, family, vivaui.HitOptions{Role: vivaui.RoleButton, Selected: selected, Cursor: vivaui.PointerCursorHand})
					if selected {
						row = row.Background(theme.Accent)
					}
					row.Children(func() {
						f.Text(family).FontSize(13).Color(theme.AccentForeground)
					})
					if row.Clicked() {
						m.selectFontFamily(family)
					}
				}
				if !found {
					f.Text("没有匹配的字体").FontSize(13).Muted().Pad(12)
				}
			})
		})
	})
	if dialog.Dismissed() {
		m.fontPickerOpen = false
	}
}

func (m *Model) selectFontFamily(family string) {
	if m.fontPickerTarget == 0 {
		switch family {
		case "Inter":
			m.settingsDraft.UIFont = "inter"
		case "系统默认":
			m.settingsDraft.UIFont = "system"
		default:
			m.settingsDraft.UIFont = "family:" + family
		}
	} else {
		switch family {
		case "JetBrains Mono":
			m.settingsDraft.CodeFont = "jetbrains"
		case "系统等宽":
			m.settingsDraft.CodeFont = "system"
		default:
			m.settingsDraft.CodeFont = "family:" + family
		}
	}
	m.fontPickerOpen = false
}

func (m *Model) fontDisplayFamily(target int) string {
	if target == 0 {
		switch m.settingsDraft.UIFont {
		case "inter":
			return "Inter"
		case "system":
			return "系统默认"
		default:
			return strings.TrimPrefix(m.settingsDraft.UIFont, "family:")
		}
	}
	switch m.settingsDraft.CodeFont {
	case "jetbrains":
		return "JetBrains Mono"
	case "system":
		return "系统等宽"
	default:
		return strings.TrimPrefix(m.settingsDraft.CodeFont, "family:")
	}
}
