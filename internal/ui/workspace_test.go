package ui

import (
	"errors"
	"fmt"
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
	if config.WindowControls == nil || config.WindowControls().CloseName != "关闭" {
		t.Fatal("custom window controls are not configured")
	}
}

func TestWorkspace(t *testing.T) {
	driver := newDriver(t, &Model{})
	defer driver.Close()

	for _, want := range []string{"JarLens", "Home", "打开...", "Ctrl+O", "拖放到此处", "就绪"} {
		if !shows(driver, want) {
			t.Errorf("home does not show %q", want)
		}
	}
	if !hasHit(driver, homeOpen, "打开...") {
		t.Error("home does not expose its open action")
	}
	if hasHit(driver, mainMenuTitle, "JarLens") {
		t.Error("home renders the archive title dropdown")
	}
	if hasHit(driver, projectTitle, "项目") || shows(driver, "尚未打开归档") {
		t.Error("home renders the project tool window or editor welcome")
	}
}

func TestHomeOpensRecentProject(t *testing.T) {
	path := writeWorkflowJAR(t)
	maximizes := 0
	m := &Model{recentProjects: []string{path}, MaximizeWindow: func() { maximizes++ }}
	driver := newDriver(t, m)
	defer driver.Close()

	pathAccessible := false
	for _, hit := range driver.Snapshot().Hits {
		pathAccessible = pathAccessible || hit.ID == homeRecentProjects.At(0) && hit.Hint == path
	}
	if !shows(driver, "最近项目") || !hasHit(driver, homeRecentProjects.At(0), "fixture.jar") || !pathAccessible {
		t.Fatal("home does not show the recent archive and its path")
	}
	click(t, driver, homeRecentProjects.At(0))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	if m.workspace.Archive.Path != path || !hasHit(driver, projectTitle, "项目") {
		t.Fatal("selecting a recent archive did not enter the full workspace")
	}
	if hasHit(driver, homeOpen, "打开...") || shows(driver, "最近项目") {
		t.Fatal("home remains visible after opening an archive")
	}
	if maximizes != 1 {
		t.Fatalf("recent project maximized the window %d times, want 1", maximizes)
	}
}

func TestHomeSecondProjectSwitchDoesNotPanic(t *testing.T) {
	m := &Model{engine: &fakeEngine{}, MaximizeWindow: func() {}, RestoreWindow: func() {}}
	driver := newDriver(t, m)
	defer driver.Close()
	first := writeWorkflowJAR(t)
	second := writeWorkflowJAR(t)
	m.recentProjects = []string{second}
	m.OpenPath(first)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })

	click(t, driver, mainMenuTitle)
	click(t, driver, mainMenuHome)
	if !m.homeVisible || !hasHit(driver, homeRecentProjects.At(1), "fixture.jar") {
		t.Fatal("Home does not list the second project after returning")
	}
	click(t, driver, homeRecentProjects.At(1))
	pumpUntil(t, driver, func() bool { return m.workspace != nil && m.workspace.Archive.Path == second })
	if m.homeVisible || !hasHit(driver, projectTitle, "项目") {
		t.Fatal("second project did not replace Home with its workbench")
	}
	click(t, driver, mainMenuTitle)
	click(t, driver, mainMenuHome)
	click(t, driver, homeRecentProjects.At(1))
	pumpUntil(t, driver, func() bool { return m.workspace != nil && m.workspace.Archive.Path == first })
	if m.homeVisible || !hasHit(driver, projectTitle, "项目") {
		t.Fatal("switching back to the first project did not restore its workbench")
	}
}

func TestMainMenuReturnsHomeAndRestoresWorkspace(t *testing.T) {
	maximizes, restores := 0, 0
	m := &Model{engine: &fakeEngine{}, MaximizeWindow: func() { maximizes++ }, RestoreWindow: func() { restores++ }}
	driver := newDriver(t, m)
	defer driver.Close()
	m.OpenPath(writeWorkflowJAR(t))
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	opened := m.workspace

	click(t, driver, mainMenuTitle)
	if !hasHit(driver, mainMenuHome, "JarLens Home") {
		t.Fatal("project menu does not expose the Home action")
	}
	click(t, driver, mainMenuHome)
	if !m.homeVisible || m.workspace != opened || !hasHit(driver, homeOpen, "打开...") || hasHit(driver, projectTitle, "项目") {
		t.Fatal("Home did not replace the workbench while retaining the current project")
	}
	if maximizes != 1 || restores != 1 {
		t.Fatalf("Home window transitions = %d maximizes, %d restores; want 1 each", maximizes, restores)
	}
	click(t, driver, mainMenuButton)
	click(t, driver, mainMenuCurrentProject)
	if m.homeVisible || m.workspace != opened || !hasHit(driver, projectTitle, "项目") || hasHit(driver, homeOpen, "打开...") {
		t.Fatal("current project did not restore the retained workbench")
	}
	if maximizes != 2 || restores != 1 {
		t.Fatalf("returning to the current project did not maximize the Home window: %d maximizes, %d restores", maximizes, restores)
	}
}

func TestRestorePathHonorsSavedWindowState(t *testing.T) {
	for _, maximize := range []bool{false, true} {
		t.Run(fmt.Sprintf("maximized=%v", maximize), func(t *testing.T) {
			maximizes := 0
			m := &Model{engine: &fakeEngine{}, MaximizeWindow: func() { maximizes++ }}
			driver := newDriver(t, m)
			defer driver.Close()
			m.RestorePath(writeWorkflowJAR(t), maximize)
			pumpUntil(t, driver, func() bool { return m.workspace != nil })
			want := 0
			if maximize {
				want = 1
			}
			if maximizes != want || m.homeVisible || !hasHit(driver, projectTitle, "项目") {
				t.Fatalf("startup restore maximizes = %d, want %d", maximizes, want)
			}
		})
	}
}

func TestOpenIDEAWorkbenchGeometry(t *testing.T) {
	maximizes := 0
	m := &Model{MaximizeWindow: func() { maximizes++ }}
	driver := newDriver(t, m)
	defer driver.Close()
	if err := driver.QueueOpenFiles([]string{writeWorkflowJAR(t)}, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, homeOpen)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	if maximizes != 1 {
		t.Fatalf("Home file selection maximized the window %d times, want 1", maximizes)
	}
	snapshot := driver.Snapshot()
	leading, _ := app.TitlebarInsets()
	wantX := map[vivaui.ID]float32{
		mainMenuButton: leading + 40,
		mainMenuTitle:  leading + 80,
		projectTitle:   14,
	}
	wantSizes := map[vivaui.ID]vivaui.Size{
		mainMenuButton:  {Width: 30, Height: 30},
		projectSearch:   {Width: 20, Height: 20},
		projectSplitter: {Width: 6},
	}
	for _, hit := range snapshot.Hits {
		if want, ok := wantX[hit.ID]; ok && hit.Bounds.X != want {
			t.Errorf("control %v starts at %v, want %v", hit.ID, hit.Bounds.X, want)
		}
		want, ok := wantSizes[hit.ID]
		if !ok {
			continue
		}
		if want.Width > 0 && hit.Bounds.Width != want.Width || want.Height > 0 && hit.Bounds.Height != want.Height {
			t.Errorf("control %v has size %vx%v, want %vx%v", hit.ID, hit.Bounds.Width, hit.Bounds.Height, want.Width, want.Height)
		}
	}
	panelPainted := false
	for _, primitive := range snapshot.Primitives {
		if primitive.Bounds.Width == 300 && primitive.Radius == 10 {
			panelPainted = true
			if primitive.Bounds.X < 6 {
				t.Errorf("project panel is flush with the window edge: %+v", primitive.Bounds)
			}
		}
	}
	if !panelPainted {
		t.Fatal("project panel was not painted")
	}
	centers := make(map[vivaui.ID]float32)
	for _, hit := range snapshot.Hits {
		switch hit.ID {
		case mainMenuButton, mainMenuTitle, projectTitle, projectSearch:
			centers[hit.ID] = hit.Bounds.Y + hit.Bounds.Height/2
		}
	}
	for _, pair := range [][2]vivaui.ID{{mainMenuButton, mainMenuTitle}, {projectTitle, projectSearch}} {
		if centers[pair[0]] != centers[pair[1]] {
			t.Errorf("controls %v and %v are not vertically centered", pair[0], pair[1])
		}
	}
	for _, label := range snapshot.Labels {
		switch label.Value {
		case "JarLens":
			if label.FontSize != 13 || label.Weight != vivaui.WeightSemibold {
				t.Errorf("archive title typography is size %v weight %v", label.FontSize, label.Weight)
			}
		case "项目":
			if label.FontSize != 13 || label.Weight != vivaui.WeightRegular {
				t.Errorf("project title typography is size %v weight %v", label.FontSize, label.Weight)
			}
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
	if !hasHit(driver, themeToggle, "使用深色主题") {
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

func TestMainMenuActions(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()

	if m.appIcon.Width != 512 || m.appIcon.Height != 512 || len(m.appIcon.Pix) == 0 {
		t.Fatal("embedded application icon was not loaded into the workspace")
	}
	if !hasHit(driver, mainMenuButton, "主菜单") || hasHit(driver, mainMenuTitle, "JarLens") {
		t.Fatal("Home titlebar must expose only the main menu, not the archive dropdown")
	}
	if err := driver.QueueOpenFiles([]string{writeWorkflowJAR(t)}, nil); err != nil {
		t.Fatal(err)
	}
	click(t, driver, mainMenuButton)
	if !shows(driver, "打开...") {
		t.Fatal("project menu does not show the existing open action")
	}
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID != mainMenuOpenFile {
			continue
		}
		for _, primitive := range driver.Snapshot().Primitives {
			if primitive.Bounds == hit.Bounds && primitive.Color == WorkspaceTheme(m.Theme).Accent {
				t.Fatal("open action has a selected background before interaction")
			}
		}
	}
	click(t, driver, mainMenuOpenFile)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	if !shows(driver, "fixture.jar") {
		t.Fatal("main-menu Open did not open the archive")
	}
	click(t, driver, mainMenuTitle)
	if !shows(driver, "当前项目") || !hasHit(driver, mainMenuCurrentProject, "fixture.jar") || !shows(driver, "F") {
		t.Fatal("project menu does not show the current archive and its badge")
	}
	var currentBounds vivaui.Rect
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == mainMenuCurrentProject {
			currentBounds = hit.Bounds
		}
	}
	for _, primitive := range driver.Snapshot().Primitives {
		if primitive.Bounds == currentBounds && primitive.Color == WorkspaceTheme(m.Theme).Accent && primitive.Radius == 5 {
			t.Fatal("current project has a selected background before interaction")
		}
	}
	if err := driver.Key(vivaui.KeyEscape, 0); err != nil {
		t.Fatal(err)
	}
	if m.mainMenuOpen {
		t.Fatal("Escape did not close the archive menu")
	}
}

func TestRecentProjectSwitchesArchive(t *testing.T) {
	m := &Model{}
	driver := newDriver(t, m)
	defer driver.Close()
	first := writeWorkflowJAR(t)
	m.OpenPath(first)
	pumpUntil(t, driver, func() bool { return m.workspace != nil })
	second := writeWorkflowJAR(t)
	m.OpenPath(second)
	pumpUntil(t, driver, func() bool { return m.workspace != nil && m.workspace.Archive.Path == second })
	if len(m.recentProjects) != 2 || m.recentProjects[0] != second || m.recentProjects[1] != first {
		t.Fatalf("recent project order is wrong: %q", m.recentProjects)
	}

	click(t, driver, mainMenuTitle)
	if !shows(driver, "最近项目") || !hasHit(driver, recentProjects.At(0), "fixture.jar") {
		t.Fatal("project menu does not list the previous archive")
	}
	click(t, driver, recentProjects.At(0))
	pumpUntil(t, driver, func() bool { return m.workspace != nil && m.workspace.Archive.Path == first })
	if len(m.documents) != 0 || m.mainMenuOpen {
		t.Fatal("recent-project switch retained stale documents or menu state")
	}
}

func TestRepositoryLink(t *testing.T) {
	for _, test := range []struct {
		name       string
		openError  error
		wantStatus string
	}{
		{name: "success", wantStatus: "正在打开 GitHub 仓库"},
		{name: "failure", openError: errors.New("unavailable"), wantStatus: "无法打开 GitHub 仓库"},
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
		{id: githubOpen, label: "打开 GitHub 仓库"},
		{id: themeToggle, label: "使用浅色主题"},
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
	if dark.SecondaryForeground != vivaui.RGB24(0x9fa2a8) || light.SecondaryForeground != vivaui.RGB24(0x5f6269) {
		t.Fatal("toolbar icons do not use the Islands secondary foreground")
	}
	foreground, gutter, selection, _ := editorColors(AppearanceDark)
	if foreground != vivaui.RGB24(0xbcbec4) || gutter != vivaui.RGB24(0x4b5059) || selection != vivaui.RGB24(0x214283) {
		t.Fatal("dark editor colors do not match the Islands editor scheme")
	}
	foreground, gutter, selection, _ = editorColors(AppearanceLight)
	if foreground != vivaui.RGB24(0x080808) || gutter != vivaui.RGB24(0xaeb3c2) || selection != vivaui.RGB24(0xa6d2ff) {
		t.Fatal("light editor colors do not match the Islands editor scheme")
	}
	if editorBannerBackground(AppearanceLight) != vivaui.RGB24(0xf7f8ff) {
		t.Fatal("light editor banner does not use its dedicated surface")
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
