package workspace

import (
	"path/filepath"
	"testing"
)

func TestSettingsStorePersistsPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := NewSettingsStoreAt(path)
	if got := store.Load(); got != DefaultSettings() {
		t.Fatalf("default settings = %+v", got)
	}
	want := Settings{Theme: "light", JavaPath: filepath.Join("opt", "jdk", "bin", "java"), Engine: "procyon", FontSize: 16, UIFont: "inter", CodeFont: "jetbrains", RestoreLastProject: false}
	store.Save(want)
	if got := NewSettingsStoreAt(path).Load(); got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}
