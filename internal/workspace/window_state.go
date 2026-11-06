package workspace

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
)

type WindowState struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Maximized bool    `json:"maximized"`
}

type WindowStore struct {
	path string
}

func NewWindowStore() *WindowStore {
	directory, err := os.UserConfigDir()
	if err != nil {
		return &WindowStore{}
	}
	return &WindowStore{path: filepath.Join(directory, "JarLens", "window.json")}
}

func NewWindowStoreAt(path string) *WindowStore {
	return &WindowStore{path: path}
}

func (store *WindowStore) Load() WindowState {
	if store == nil || store.path == "" {
		return WindowState{}
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		return WindowState{}
	}
	var state WindowState
	if json.Unmarshal(data, &state) != nil || !state.valid() {
		return WindowState{}
	}
	return state
}

func (store *WindowStore) Save(state WindowState) {
	if store == nil || store.path == "" || !state.valid() {
		return
	}
	data, err := json.Marshal(state)
	if err != nil || os.MkdirAll(filepath.Dir(store.path), 0o700) != nil {
		return
	}
	temporary, err := os.CreateTemp(filepath.Dir(store.path), "window-*.json")
	if err != nil {
		return
	}
	name := temporary.Name()
	_, writeErr := temporary.Write(data)
	closeErr := temporary.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(name)
		return
	}
	if os.Rename(name, store.path) != nil {
		_ = os.Remove(name)
	}
}

func (state WindowState) valid() bool {
	return state.Width > 0 && state.Height > 0 && isFinite(state.X) && isFinite(state.Y)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
