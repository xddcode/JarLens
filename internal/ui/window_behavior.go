package ui

import "github.com/vivaui/viva/app"

func (m *Model) ConfigureWindow(application *app.Application, window *app.Window) {
	m.application = application
	m.window = window
}

func (m *Model) hideToTray() bool {
	if m.HideWindow == nil && m.window == nil {
		return false
	}
	if !m.ensureStatusItem() {
		return false
	}
	if m.HideWindow != nil {
		m.HideWindow()
	} else {
		m.window.Hide()
	}
	return true
}

func (m *Model) ensureStatusItem() bool {
	if m.EnsureTray != nil {
		return m.EnsureTray()
	}
	if m.statusItem != nil {
		return true
	}
	if m.application == nil {
		return false
	}
	item, err := m.application.AddStatusItem(app.StatusItemConfig{
		Tooltip: "JarLens",
		Icon:    m.TrayIcon.StatusIcon(16),
		OnClick: func(app.StatusClick) {
			if m.window != nil {
				m.window.Show()
			}
		},
		Menu: func() []app.MenuItem {
			return []app.MenuItem{
				{Title: m.t("打开 JarLens"), OnSelect: func() {
					if m.window != nil {
						m.window.Show()
					}
				}},
				{Title: m.t("退出"), OnSelect: m.application.Quit},
			}
		},
	})
	if err != nil {
		return false
	}
	m.statusItem = item
	return true
}
