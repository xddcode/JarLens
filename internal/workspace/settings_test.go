package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsStorePersistsPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := NewSettingsStoreAt(path)
	if got := store.Load(); got != DefaultSettings() {
		t.Fatalf("default settings = %+v", got)
	}
	want := Settings{Theme: "light", Language: "zh-TW", JavaPath: filepath.Join("opt", "jdk", "bin", "java"), Engine: "procyon", FontSize: 16, UIFont: "inter", CodeFont: "jetbrains", TabWidth: 6, RestoreLastProject: false, CloseToTray: true, LaunchAtLogin: true, UpdateSource: "cnb"}
	store.Save(want)
	if got := NewSettingsStoreAt(path).Load(); got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestSettingsStorePersistsSystemTheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	want := DefaultSettings()
	want.Theme = "system"
	NewSettingsStoreAt(path).Save(want)
	if got := NewSettingsStoreAt(path).Load(); got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestSettingsStoreDefaultsIndentGuidesForLegacySettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"dark"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := NewSettingsStoreAt(path).Load(); !got.ShowIndentGuides {
		t.Fatal("legacy settings disabled indent guides by default")
	}
}
