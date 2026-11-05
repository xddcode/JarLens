package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/vivaui/viva/app"
	vivaui "github.com/vivaui/viva/ui"
)

func TestConfig(t *testing.T) {
	config := NewConfig(&Model{})
	if config.Title != "JarLens" {
		t.Fatalf("window title is %q", config.Title)
	}
	if config.Width < 900 || config.Height < 600 {
		t.Fatalf("workspace window is too small: %dx%d", config.Width, config.Height)
	}
	if config.WindowMode != app.WindowCustomTitlebar || config.TitlebarHeight != 40 {
		t.Fatal("workspace does not use the integrated 40px titlebar")
	}
	if config.WindowControls == nil || config.WindowControls().CloseName != "Close" {
		t.Fatal("custom window controls are not configured")
	}
}

func TestWorkspace(t *testing.T) {
	driver := newDriver(t, &Model{})
	defer driver.Close()

	for _, want := range []string{"JarLens", "Project", "No archive open", "Explore Java bytecode.", "Ready"} {
		if !shows(driver, want) {
			t.Errorf("workspace does not show %q", want)
		}
	}
}

func TestThemeSwitch(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()

	if len(driver.Snapshot().Primitives) == 0 {
		t.Fatal("workspace painted no background")
	}
	dark := driver.Snapshot().Primitives[0].Color
	click(t, driver, themeToggle)
	if m.Theme != AppearanceLight {
		t.Fatalf("theme after switching to light is %v", m.Theme)
	}
	if !hasHit(driver, themeToggle, "Use dark theme") {
		t.Error("theme control does not offer the dark theme")
	}
	if len(driver.Snapshot().Primitives) == 0 {
		t.Fatal("light workspace painted no background")
	}
	if light := driver.Snapshot().Primitives[0].Color; light == dark {
		t.Error("switching themes did not change the painted window background")
	}

	click(t, driver, themeToggle)
	if m.Theme != AppearanceDark {
		t.Fatalf("theme after switching back to dark is %v", m.Theme)
	}
}

func TestRepositoryLink(t *testing.T) {
	for _, test := range []struct {
		name       string
		openError  error
		wantStatus string
	}{
		{name: "success", wantStatus: "Opening GitHub repository"},
		{name: "failure", openError: errors.New("unavailable"), wantStatus: "Could not open GitHub repository"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var opened string
			m := &Model{openURL: func(target string) error {
				opened = target
				return test.openError
			}}
			driver := newDriver(t, m)
			defer driver.Close()

			click(t, driver, githubOpen)
			if opened != repositoryURL {
				t.Fatalf("opened %q instead of the JarLens repository", opened)
			}
			if !shows(driver, test.wantStatus) {
				t.Errorf("workspace does not show %q", test.wantStatus)
			}
		})
	}
}

func TestIconButtonsHaveAccessibleNamesAndHints(t *testing.T) {
	driver := newDriver(t, &Model{})
	defer driver.Close()

	for _, test := range []struct {
		id    vivaui.ID
		label string
	}{
		{id: githubOpen, label: "Open GitHub repository"},
		{id: themeToggle, label: "Use light theme"},
	} {
		if !hasHit(driver, test.id, test.label) {
			t.Errorf("missing accessible icon button %q", test.label)
		}
		for _, hit := range driver.Snapshot().Hits {
			if hit.ID == test.id && (hit.Hint != test.label || hit.Role != vivaui.RoleButton) {
				t.Errorf("icon button %q lacks its hint or button role", test.label)
			}
		}
	}
}

func TestWorkspaceLabelsStayWithinWindow(t *testing.T) {
	for _, appearance := range []Appearance{AppearanceDark, AppearanceLight} {
		t.Run(appearanceName(appearance), func(t *testing.T) {
			driver := newDriver(t, &Model{Theme: appearance})
			defer driver.Close()

			snapshot := driver.Snapshot()
			for _, label := range snapshot.Labels {
				right := label.Bounds.X + label.Bounds.Width
				bottom := label.Bounds.Y + label.Bounds.Height
				if label.Bounds.X < 0 || label.Bounds.Y < 0 || right > float32(snapshot.Width) || bottom > float32(snapshot.Height) {
					t.Errorf("label %q is outside the window: bounds %+v", label.Value, label.Bounds)
				}
			}
		})
	}
}

func TestWorkspaceThemes(t *testing.T) {
	dark := WorkspaceTheme(AppearanceDark)
	light := WorkspaceTheme(AppearanceLight)
	if dark.Background == light.Background {
		t.Fatal("dark and light themes share a background")
	}
	if dark.Primary != vivaui.RGB24(0x3871e1) || light.Primary != dark.Primary {
		t.Fatal("themes do not preserve the semantic accent")
	}
	if dark.Metrics.ControlHeight != 28 || light.Metrics.RadiusSurface != 10 {
		t.Fatal("themes do not share workspace metrics")
	}
}

func newDriver(t *testing.T, m *Model) *app.Driver {
	t.Helper()
	var canvas vivaui.Canvas
	driver, err := app.NewDriver(NewConfig(m), &canvas)
	if err != nil {
		t.Fatal(err)
	}
	return driver
}

func click(t *testing.T, driver *app.Driver, id vivaui.ID) {
	t.Helper()
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == id {
			bounds := hit.Bounds
			if err := driver.Click(bounds.X+bounds.Width/2, bounds.Y+bounds.Height/2); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("no control %v on screen", id)
}

func hasHit(driver *app.Driver, id vivaui.ID, label string) bool {
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == id && hit.Label == label {
			return true
		}
	}
	return false
}

func shows(driver *app.Driver, want string) bool {
	for _, label := range driver.Snapshot().Labels {
		if strings.Contains(label.Value, want) {
			return true
		}
	}
	return false
}

func appearanceName(appearance Appearance) string {
	if appearance == AppearanceLight {
		return "light"
	}
	return "dark"
}
