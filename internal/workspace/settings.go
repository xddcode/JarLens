package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	Theme                string `json:"theme"`
	Language             string `json:"language"`
	JavaPath             string `json:"javaPath"`
	Engine               string `json:"engine"`
	FontSize             int    `json:"fontSize"`
	UIFont               string `json:"uiFont"`
	CodeFont             string `json:"codeFont"`
	TabWidth             int    `json:"tabWidth"`
	ShowIndentGuides     bool   `json:"showIndentGuides"`
	ShowLineNumbers      bool   `json:"showLineNumbers"`
	HighlightCurrentLine bool   `json:"highlightCurrentLine"`
	RestoreLastProject   bool   `json:"restoreLastProject"`
	CheckUpdates         bool   `json:"checkUpdates"`
	UpdateSource         string `json:"updateSource"`
}

func DefaultSettings() Settings {
	return Settings{Theme: "dark", Language: "zh-CN", Engine: "vineflower", FontSize: 13, UIFont: "inter", CodeFont: "jetbrains", TabWidth: 4, ShowIndentGuides: true, ShowLineNumbers: true, HighlightCurrentLine: true, RestoreLastProject: true, CheckUpdates: true, UpdateSource: "github"}
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
	var saved struct {
		Settings
		ShowIndentGuides *bool `json:"showIndentGuides"`
		CheckUpdates     *bool `json:"checkUpdates"`
	}
	if json.Unmarshal(data, &saved) != nil {
		return settings
	}
	if saved.Theme == "light" || saved.Theme == "dark" || saved.Theme == "system" {
		settings.Theme = saved.Theme
	}
	if saved.Language == "zh-CN" || saved.Language == "en" || saved.Language == "ja" || saved.Language == "ko" {
		settings.Language = saved.Language
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
	if saved.TabWidth == 2 || saved.TabWidth == 4 || saved.TabWidth == 6 {
		settings.TabWidth = saved.TabWidth
	}
	if saved.ShowIndentGuides != nil {
		settings.ShowIndentGuides = *saved.ShowIndentGuides
	}
	settings.ShowLineNumbers = saved.ShowLineNumbers
	settings.HighlightCurrentLine = saved.HighlightCurrentLine
	settings.RestoreLastProject = saved.RestoreLastProject
	if saved.CheckUpdates != nil {
		settings.CheckUpdates = *saved.CheckUpdates
	}
	if saved.UpdateSource == "github" || saved.UpdateSource == "cnb" {
		settings.UpdateSource = saved.UpdateSource
	}
	if saved.Engine == "vineflower" || saved.Engine == "cfr" || saved.Engine == "procyon" {
		settings.Engine = saved.Engine
	}
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
