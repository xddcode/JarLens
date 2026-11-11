package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLaunchItemFileReplacementAndRemoval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "autostart", "jarlens.desktop")
	if err := writeLaunchItem(path, []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := writeLaunchItem(path, []byte("second")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "second" {
		t.Fatalf("launch item = %q, want replacement", data)
	}
	if err := removeLaunchItem(path); err != nil {
		t.Fatal(err)
	}
	if err := removeLaunchItem(path); err != nil {
		t.Fatalf("removing an absent launch item failed: %v", err)
	}
}

func TestLaunchItemFileReportsWriteErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(path, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeLaunchItem(filepath.Join(path, "jarlens.desktop"), []byte("data")); err == nil {
		t.Fatal("launch item write succeeded with an invalid parent directory")
	}
}
