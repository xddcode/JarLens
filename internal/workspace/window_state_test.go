package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowStorePersistsPlacementAndMaximizedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")
	store := NewWindowStoreAt(path)
	want := WindowState{X: 120, Y: 80, Width: 1440, Height: 900, Maximized: true}
	store.Save(want)
	if got := NewWindowStoreAt(path).Load(); got != want {
		t.Fatalf("window state = %+v, want %+v", got, want)
	}
}

func TestWindowStoreRejectsInvalidState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window.json")
	if err := os.WriteFile(path, []byte(`{"width":0,"height":760}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := NewWindowStoreAt(path).Load(); got != (WindowState{}) {
		t.Fatalf("invalid window state was loaded: %+v", got)
	}
}
