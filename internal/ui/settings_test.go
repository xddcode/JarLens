package ui

import (
	"os"
	"testing"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/ui"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

type memorySettingsStore struct {
	settings jarlensworkspace.Settings
	saves    int
}

func (store *memorySettingsStore) Load() jarlensworkspace.Settings {
	return store.settings
}

func (store *memorySettingsStore) Save(settings jarlensworkspace.Settings) {
	store.settings = settings
	store.saves++
}

func openSettingsForTest(t *testing.T, driver *app.Driver) {
	t.Helper()
	handled, err := driver.ShortcutChord(ui.ModifierControl|ui.ModifierAlt, 's')
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("settings shortcut was not handled")
	}
}

func selectSettingsTheme(t *testing.T, driver *app.Driver, index int) {
	t.Helper()
	click(t, driver, settingsTheme.At(3))
	click(t, driver, settingsTheme.At(index))
}

func TestSettingsDialogAppliesAndPersists(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store, engine: &fakeEngine{}}
	driver := newDriver(t, m)
	defer driver.Close()

	openSettingsForTest(t, driver)
	if !m.settingsOpen || !shows(driver, "外观与行为") || !shows(driver, "编辑器") || !shows(driver, "系统设置") {
		t.Fatal("settings dialog does not show its page tree")
	}
	selectSettingsTheme(t, driver, 1)
	click(t, driver, settingsNav.At(3))
	m.settingsFontSize.SetText("16")
	click(t, driver, settingsNav.At(5))
	click(t, driver, settingsRestore)
	java, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	m.settingsJava.SetText(java)
	click(t, driver, settingsConfirm)
	if m.settingsOpen || m.Theme != AppearanceLight || m.settings.FontSize != 16 || m.settings.RestoreLastProject || m.settings.JavaPath != java {
		t.Fatalf("settings were not applied: %+v", m.settings)
	}
	if store.saves != 1 || store.settings != m.settings {
		t.Fatal("settings were not persisted")
	}

	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(1))
	selectSettingsTheme(t, driver, 0)
	click(t, driver, settingsCancel)
	if m.settingsOpen || m.Theme != AppearanceLight || store.saves != 1 {
		t.Fatalf("cancel changed state: open=%v theme=%v saves=%d", m.settingsOpen, m.Theme, store.saves)
	}
}

func TestSettingsEditorAndSystemPages(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(3))
	click(t, driver, settingsCodeFontChoose)
	click(t, driver, settingsCodeFont.At(1))
	click(t, driver, settingsNav.At(4))
	click(t, driver, settingsTabWidth.At(0))
	click(t, driver, settingsIndentGuides)
	click(t, driver, settingsLineNumbers)
	m.view.ScrollState(settingsContentScroll).ScrollToEnd()
	if err := driver.Move(0, 0); err != nil {
		t.Fatal(err)
	}
	click(t, driver, settingsCurrentLine)
	click(t, driver, settingsNav.At(5))
	click(t, driver, settingsRestore)
	click(t, driver, settingsApply)
	if !m.settingsOpen || store.saves != 1 {
		t.Fatal("Apply did not save settings while keeping the dialog open")
	}
	if m.settings.CodeFont != "system" || m.settings.ShowIndentGuides || m.settings.ShowLineNumbers || m.settings.HighlightCurrentLine || m.settings.RestoreLastProject {
		t.Fatalf("editor or system settings were not applied: %+v", m.settings)
	}
	if m.settingsDirty() {
		t.Fatal("applied settings are still marked as modified")
	}
	click(t, driver, settingsCancel)
	if m.settings.CodeFont != "system" || store.saves != 1 {
		t.Fatal("Cancel reverted settings that were already applied")
	}
}

func TestSettingsApplyOnlyWhenDirty(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)

	click(t, driver, settingsApply)
	if store.saves != 0 || !m.settingsOpen {
		t.Fatal("Apply was enabled without staged changes")
	}
	selectSettingsTheme(t, driver, 1)
	click(t, driver, settingsApply)
	if store.saves != 1 || m.settings.Theme != "light" || !m.settingsOpen {
		t.Fatal("Apply did not save the staged theme")
	}
	click(t, driver, settingsApply)
	if store.saves != 1 {
		t.Fatal("Apply saved unchanged settings more than once")
	}
}

func TestSettingsSelectAndPersistDecompilerEngine(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(5))
	click(t, driver, settingsEngine.At(3))
	click(t, driver, settingsEngine.At(1))
	click(t, driver, settingsApply)
	if store.saves != 1 || m.settings.Engine != "cfr" || m.engine.Name() != "CFR" {
		t.Fatalf("CFR was not selected and persisted: %+v", m.settings)
	}

	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(5))
	click(t, driver, settingsEngine.At(3))
	click(t, driver, settingsEngine.At(2))
	click(t, driver, settingsCancel)
	if m.settings.Engine != "cfr" || m.engine.Name() != "CFR" {
		t.Fatal("Cancel changed the applied decompiler engine")
	}
}

func TestSettingsThemeSelectRemainsVisibleDuringEntrance(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsTheme.At(3))

	for _, interval := range []time.Duration{0, 16 * time.Millisecond, 64 * time.Millisecond, 200 * time.Millisecond} {
		if interval > 0 {
			if err := driver.Advance(interval); err != nil {
				t.Fatal(err)
			}
		}
		if !m.settingsThemeSelect.Open || !m.settingsOpen {
			t.Fatalf("select closed during its entrance after %v", interval)
		}
		if !hasHit(driver, settingsTheme.At(0), "深色") || !hasHit(driver, settingsTheme.At(1), "浅色") || !hasHit(driver, settingsTheme.At(2), "跟随系统") {
			t.Fatalf("select options disappeared during its entrance after %v", interval)
		}
		if !shows(driver, "设置") {
			t.Fatalf("settings dialog disappeared during select entrance after %v", interval)
		}
	}
	click(t, driver, settingsTheme.At(1))
	if m.settingsThemeSelect.Open || m.settingsDraft.Theme != "light" {
		t.Fatal("select did not apply the chosen theme and close")
	}
}

func TestSettingsFollowSystemAppearance(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	driver.SetSystemAppearance(app.AppearanceLight)

	openSettingsForTest(t, driver)
	selectSettingsTheme(t, driver, 2)
	click(t, driver, settingsApply)
	if store.saves != 1 || m.settings.Theme != "system" || m.Theme != AppearanceLight {
		t.Fatalf("system theme was not applied and persisted: settings=%+v appearance=%v", m.settings, m.Theme)
	}
	if got := driver.NativeAppearance(); got != app.AppearanceUnknown {
		t.Fatalf("native appearance = %v, want system-controlled", got)
	}

	driver.SetSystemAppearance(app.AppearanceDark)
	pumpUntil(t, driver, func() bool { return m.Theme == AppearanceDark })
	if m.settings.Theme != "system" {
		t.Fatal("system appearance change replaced the saved preference")
	}
}

func TestSettingsSearchNavigatesToSetting(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsSearch)
	if err := driver.Type("字号"); err != nil {
		t.Fatal(err)
	}
	if !hasHit(driver, settingsNav.At(3), "字体") || hasHit(driver, settingsNav.At(1), "外观") {
		t.Fatal("settings search did not filter the page tree")
	}
	if m.settingsCategory != 3 {
		t.Fatalf("settings search opened page %d, want editor font", m.settingsCategory)
	}
}

func TestSettingsSearchKeyboard(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsSearch)
	if err := driver.Type("字号"); err != nil {
		t.Fatal(err)
	}
	if err := driver.Key(ui.KeyEscape, 0); err != nil {
		t.Fatal(err)
	}
	if !m.settingsOpen || m.settingsSearch.Text() != "" {
		t.Fatal("Escape closed the dialog instead of clearing settings search")
	}
	before := m.settingsCategory
	if err := driver.Key(ui.KeyDown, 0); err != nil {
		t.Fatal(err)
	}
	if m.settingsCategory == before {
		t.Fatal("Arrow Down did not select the next settings page")
	}
	if focused := driver.Snapshot().Focused; focused != settingsSearch {
		t.Fatalf("settings search lost keyboard focus: %v", focused)
	}
}

func TestSettingsHistoryAndResetCurrentPage(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	selectSettingsTheme(t, driver, 1)
	click(t, driver, settingsNav.At(3))
	m.settingsFontSize.SetText("16")
	m.settingsDraft.FontSize = 16
	click(t, driver, settingsNav.At(5))
	click(t, driver, settingsBack)
	if m.settingsCategory != 3 {
		t.Fatalf("Back opened page %d, want font", m.settingsCategory)
	}
	click(t, driver, settingsForward)
	if m.settingsCategory != 5 {
		t.Fatalf("Forward opened page %d, want system", m.settingsCategory)
	}
	click(t, driver, settingsBack)
	click(t, driver, settingsReset)
	if m.settingsFontSizeValue() != 13 || m.settingsDraft.Theme != "light" {
		t.Fatal("Reset changed another page or did not restore the font page")
	}
	click(t, driver, settingsNav.At(1))
	click(t, driver, settingsReset)
	if m.settingsDraft.Theme != "dark" || m.settingsDirty() {
		t.Fatal("Reset did not discard the remaining staged appearance change")
	}
}

func TestSettingsGroupPageLinks(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(0))
	if !shows(driver, "自定义 JarLens 的主题") {
		t.Fatal("appearance group page is missing its description")
	}
	click(t, driver, settingsGroupLinks.At(5))
	if m.settingsCategory != 5 {
		t.Fatalf("group link opened page %d, want system settings", m.settingsCategory)
	}
}

func TestSettingsRejectsInvalidFontSize(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(3))
	m.settingsFontSize.SetText("99")
	click(t, driver, settingsConfirm)
	if !m.settingsOpen || store.saves != 0 || !shows(driver, "字号应为 10 到 24 的整数") {
		t.Fatal("invalid editor font size was accepted")
	}
	click(t, driver, settingsReset)
	if m.settingsError != "" || m.settingsFontSize.Text() != "13" {
		t.Fatal("Reset did not restore the valid font size")
	}
}

func TestSettingsFontSizeArrowKeys(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(3))
	click(t, driver, settingsFontSizeInput)
	if err := driver.Key(ui.KeyUp, 0); err != nil {
		t.Fatal(err)
	}
	if m.settingsFontSize.Text() != "14" || m.settingsDraft.FontSize != 14 {
		t.Fatalf("Arrow Up did not increment font size: %q", m.settingsFontSize.Text())
	}
}

func TestFontPickerFiltersAndSelectsFont(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(3))
	click(t, driver, settingsCodeFontChoose)
	if !m.fontPickerOpen {
		t.Fatal("font picker did not open")
	}
	var settingsLayer, pickerLayer uint16
	for _, hit := range driver.Snapshot().Hits {
		switch hit.ID {
		case settingsClose:
			settingsLayer = hit.Layer
		case fontPickerClose:
			pickerLayer = hit.Layer
		}
	}
	if pickerLayer <= settingsLayer {
		t.Fatalf("font picker layer %d is not above settings layer %d", pickerLayer, settingsLayer)
	}
	click(t, driver, fontPickerSearch)
	if err := driver.Type("等宽"); err != nil {
		t.Fatal(err)
	}
	click(t, driver, settingsCodeFont.At(1))
	if m.fontPickerOpen || m.settingsDraft.CodeFont != "system" {
		t.Fatal("font picker did not apply the selected font")
	}
}

func TestSettingsRejectsInvalidJavaPath(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	m := &Model{SettingsStore: store}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	m.settingsJava.SetText("missing-java executable")
	click(t, driver, settingsConfirm)
	if !m.settingsOpen || store.saves != 0 || !shows(driver, "Java 可执行文件路径无效") {
		t.Fatal("invalid Java path was accepted")
	}
}

func TestSettingsGlyphUsesOpenIDEAIcon(t *testing.T) {
	for _, appearance := range []Appearance{AppearanceDark, AppearanceLight} {
		glyph := settingsGlyphForAppearance(appearance)
		if glyph.ViewBox != 20 || len(glyph.Layers) != 1 {
			t.Fatalf("settings glyph does not use OpenIDEA's 20-pixel icon: %+v", glyph)
		}
		if glyph.Layers[0].FillRule != ui.PathFillEvenOdd || len(glyph.Layers[0].Commands) == 0 {
			t.Fatal("settings glyph lost the OpenIDEA gear's cutouts")
		}
	}
	if settingsGlyphForAppearance(AppearanceDark).Layers[0].Fill == settingsGlyphForAppearance(AppearanceLight).Layers[0].Fill {
		t.Fatal("settings glyph does not follow the workspace theme")
	}
}

func TestSettingsButtonOpensDialog(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	click(t, driver, settingsButton)
	if !m.settingsOpen || !shows(driver, "设置") {
		t.Fatal("settings toolbar button did not open the dialog")
	}
}

func TestSettingsShortcut(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	handled, err := driver.ShortcutChord(ui.ModifierControl|ui.ModifierAlt, 's')
	if err != nil {
		t.Fatal(err)
	}
	if !handled || !m.settingsOpen {
		t.Fatal("Ctrl+Alt+S did not open settings")
	}
}

func TestSettingsReopensLastPage(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(4))
	click(t, driver, settingsCancel)
	openSettingsForTest(t, driver)
	if m.settingsCategory != 4 {
		t.Fatalf("settings reopened page %d, want editor general", m.settingsCategory)
	}
}
