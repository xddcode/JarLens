package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecentStorePersistsRecentFirstOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recent.json")
	store := NewRecentStoreAt(path)
	first := filepath.Join("work", "first.jar")
	second := filepath.Join("work", "second.jar")
	store.Record(first)
	store.Record(second)
	store.Record(first)
	want := []string{first, second}
	if got := NewRecentStoreAt(path).Load(); !reflect.DeepEqual(got, want) {
		t.Fatalf("recent projects = %q, want %q", got, want)
	}
}

func TestRecentStoreRejectsInvalidHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recent.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := NewRecentStoreAt(path).Load(); len(got) != 0 {
		t.Fatalf("invalid history was loaded: %q", got)
	}
}

func TestRecentStoreLimitsHistory(t *testing.T) {
	store := NewRecentStoreAt(filepath.Join(t.TempDir(), "recent.json"))
	for index := 0; index < 20; index++ {
		store.Record(filepath.Join("work", string(rune('a'+index))+".jar"))
	}
	if got := len(store.Load()); got != recentLimit {
		t.Fatalf("recent project count = %d, want %d", got, recentLimit)
	}
}
