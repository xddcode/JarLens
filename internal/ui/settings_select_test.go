package ui

import (
	"fmt"
	"testing"

	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
)

func TestSettingsSelectDismissesOnAnyOutsideClick(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		trigger int
		scope   vivaui.Scope
		state   func(*Model) *flex.SelectState
	}{
		{name: "theme", page: 1, trigger: 3, scope: settingsTheme, state: func(m *Model) *flex.SelectState { return &m.settingsThemeSelect }},
		{name: "language", page: 1, trigger: 4, scope: settingsLanguage, state: func(m *Model) *flex.SelectState { return &m.settingsLanguageSelect }},
		{name: "update source", page: 6, trigger: 2, scope: settingsUpdateSource, state: func(m *Model) *flex.SelectState { return &m.settingsUpdateSourceSelect }},
		{name: "engine", page: 5, trigger: 3, scope: settingsEngine, state: func(m *Model) *flex.SelectState { return &m.settingsEngineSelect }},
	}
	for _, test := range tests {
		for _, point := range []vivaui.Point{
			{X: 20, Y: 20},
			{X: 120, Y: 180},
			{X: 1120, Y: 740},
			{X: 120, Y: 700},
		} {
			t.Run(fmt.Sprintf("%s/%v", test.name, point), func(t *testing.T) {
				m := &Model{}
				driver := newDriver(t, m)
				defer driver.Close()
				openSettingsForTest(t, driver)
				if test.page != 1 {
					click(t, driver, settingsNav.At(test.page))
				}
				click(t, driver, test.scope.At(test.trigger))
				state := test.state(m)
				if !state.Open {
					t.Fatal("select did not open")
				}
				category := m.settingsCategory
				if err := driver.Click(point.X, point.Y); err != nil {
					t.Fatal(err)
				}
				if state.Open {
					t.Errorf("select stayed open after outside click at %+v", point)
				}
				if !m.settingsOpen {
					t.Fatal("outside click closed the settings dialog instead of only the select")
				}
				if m.settingsCategory != category {
					t.Fatal("outside click activated the page beneath the select")
				}
			})
		}
	}
}
