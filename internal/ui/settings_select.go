package ui

import vivaui "github.com/vivaui/viva/ui"

func (m *Model) paintSettingsSelectDismissBackdrop(canvas *vivaui.Canvas, bounds vivaui.Rect) {
	dismiss := vivaui.ID{}
	focusScope := vivaui.ID{}
	switch {
	case m.settingsThemeSelect.Open:
		dismiss, focusScope = settingsTheme.At(4), settingsTheme.At(5)
	case m.settingsLanguageSelect.Open:
		dismiss, focusScope = settingsLanguage.At(6), settingsLanguage.At(7)
	case m.settingsUpdateSourceSelect.Open:
		dismiss, focusScope = settingsUpdateSource.At(3), settingsUpdateSource.At(4)
	case m.settingsEngineSelect.Open:
		dismiss, focusScope = settingsEngine.At(4), settingsEngine.At(5)
	}
	if dismiss.IsZero() {
		return
	}
	// The component's backdrop can be clipped by the page's scroll area. This
	// full-window hit sits above the dialog and below the select's option layer.
	canvas.PushLayer(2)
	canvas.PushFocusScope(focusScope)
	canvas.HitWithOptions(dismiss, bounds, m.t("关闭下拉"), vivaui.HitOptions{
		Role: vivaui.RoleButton, SkipFocus: true, Backdrop: true, PassSecondary: true,
	})
	canvas.PopFocusScope()
	canvas.PopLayer()
}
