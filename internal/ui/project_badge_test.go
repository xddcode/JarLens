package ui

import "testing"

func TestProjectInitials(t *testing.T) {
	for path, want := range map[string]string{
		"eco-aiot-platform-at": "EA",
		"jd-gui-1.6.6.jar":     "JG",
		"fixture.jar":          "F",
		`C:\work\fixture.jar`:  "F",
		"JarLens":              "J",
	} {
		if got := projectInitials(path); got != want {
			t.Errorf("projectInitials(%q) = %q, want %q", path, got, want)
		}
	}
	if projectBadgeColor("fixture.jar") != projectBadgeColor("fixture.jar") {
		t.Fatal("project badge color is not deterministic")
	}
	if projectBadgeColor("fixture.jar") != projectBadgeColor(`C:\work\fixture.jar`) {
		t.Fatal("the same project has different badge colors in the titlebar and menu")
	}
}
