package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	Theme                string `json:"theme"`
	JavaPath             string `json:"javaPath"`
	FontSize             int    `json:"fontSize"`
	UIFont               string `json:"uiFont"`
	CodeFont             string `json:"codeFont"`
	ShowLineNumbers      bool   `json:"showLineNumbers"`
	HighlightCurrentLine bool   `json:"highlightCurrentLine"`
	RestoreLastProject   bool   `json:"restoreLastProject"`
}

func DefaultSettings() Settings {
	return Settings{Theme: "dark", FontSize: 13, UIFont: "inter", CodeFont: "jetbrains", ShowLineNumbers: true, HighlightCurrentLine: true, RestoreLastProject: true}
}

type SettingsStore struct {
	path string
}

func NewSettingsStore() *SettingsStore {
	directory, err := os.UserConfigDir()
	if err != nil {
		return &SettingsStore{}
	}
	return &SettingsStore{path: filepath.Join(directory, "JarLens", "settings.json")}
}

func NewSettingsStoreAt(path string) *SettingsStore {
	return &SettingsStore{path: path}
}

func (store *SettingsStore) Load() Settings {
	settings := DefaultSettings()
	if store == nil || store.path == "" {
		return settings
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		return settings
	}
	var saved Settings
	if json.Unmarshal(data, &saved) != nil {
		return settings
	}
	if saved.Theme == "light" || saved.Theme == "dark" {
		settings.Theme = saved.Theme
	}
	settings.JavaPath = strings.TrimSpace(saved.JavaPath)
	if saved.FontSize >= 10 && saved.FontSize <= 24 {
		settings.FontSize = saved.FontSize
	}
	if saved.UIFont == "inter" || saved.UIFont == "system" || strings.HasPrefix(saved.UIFont, "family:") {
		settings.UIFont = saved.UIFont
	}
	if saved.CodeFont == "jetbrains" || saved.CodeFont == "system" || strings.HasPrefix(saved.CodeFont, "family:") {
		settings.CodeFont = saved.CodeFont
	}
	settings.ShowLineNumbers = saved.ShowLineNumbers
	settings.HighlightCurrentLine = saved.HighlightCurrentLine
	settings.RestoreLastProject = saved.RestoreLastProject
	return settings
}

func (store *SettingsStore) Save(settings Settings) {
	if store == nil || store.path == "" {
		return
	}
	settings.JavaPath = strings.TrimSpace(settings.JavaPath)
	data, err := json.Marshal(settings)
	if err != nil || os.MkdirAll(filepath.Dir(store.path), 0o700) != nil {
		return
	}
	temporary, err := os.CreateTemp(filepath.Dir(store.path), "settings-*.json")
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
