package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const recentLimit = 12

type RecentStore struct {
	path string
}

func NewRecentStore() *RecentStore {
	directory, err := os.UserConfigDir()
	if err != nil {
		return &RecentStore{}
	}
	return &RecentStore{path: filepath.Join(directory, "JarLens", "recent.json")}
}

func NewRecentStoreAt(path string) *RecentStore {
	return &RecentStore{path: path}
}

func (store *RecentStore) Load() []string {
	if store == nil || store.path == "" {
		return nil
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		return nil
	}
	var paths []string
	if json.Unmarshal(data, &paths) != nil {
		return nil
	}
	return normalizeRecent(paths)
}

func (store *RecentStore) Record(path string) []string {
	paths := RecordRecent(store.Load(), path)
	if store == nil || store.path == "" {
		return paths
	}
	data, err := json.Marshal(paths)
	if err != nil {
		return paths
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o700); err != nil {
		return paths
	}
	temporary, err := os.CreateTemp(filepath.Dir(store.path), "recent-*.json")
	if err != nil {
		return paths
	}
	name := temporary.Name()
	_, writeErr := temporary.Write(data)
	closeErr := temporary.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(name)
		return paths
	}
	if err := os.Rename(name, store.path); err != nil {
		_ = os.Remove(name)
	}
	return paths
}

func RecordRecent(paths []string, path string) []string {
	path = filepath.Clean(strings.TrimSpace(path))
	paths = append([]string{path}, paths...)
	paths = normalizeRecent(paths)
	if len(paths) > recentLimit {
		paths = paths[:recentLimit]
	}
	return paths
}

func normalizeRecent(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.Clean(strings.TrimSpace(path))
		if path == "" || path == "." {
			continue
		}
		duplicate := false
		for _, existing := range result {
			if existing == path || runtime.GOOS == "windows" && strings.EqualFold(existing, path) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, path)
		}
	}
	return result
}
