package main

import (
	"log"
	"os"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/app/iconpng"
	"github.com/vivaui/viva/ui"
	jarlensassets "github.com/xddcode/jarlens/assets"
	jarlensui "github.com/xddcode/jarlens/internal/ui"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

func main() {
	icon, err := iconpng.DecodeFit(jarlensassets.AppIconPNG, 512)
	if err != nil {
		log.Fatal(err)
	}
	recentStore := jarlensworkspace.NewRecentStore()
	windowStore := jarlensworkspace.NewWindowStore()
	settingsStore := jarlensworkspace.NewSettingsStore()
	preferences := settingsStore.Load()
	windowState := windowStore.Load()
	model := &jarlensui.Model{RecentStore: recentStore, SettingsStore: settingsStore}
	config := jarlensui.NewConfig(model)
	config.Icon = icon
	if windowState.Width > 0 {
		config.Width = windowState.Width
		config.Height = windowState.Height
		config.Origin = &ui.Point{X: float32(windowState.X), Y: float32(windowState.Y)}
	}
	var window *app.Window
	config.OnFrameChange = func(frame ui.Rect) {
		if window != nil && (window.Maximized() || window.Fullscreen()) {
			return
		}
		windowState.X = float64(frame.X)
		windowState.Y = float64(frame.Y)
		windowState.Width = int(frame.Width)
		windowState.Height = int(frame.Height)
	}
	config.OnMaximize = func(maximized bool) {
		windowState.Maximized = maximized
		windowStore.Save(windowState)
	}
	config.OnFullscreen = func(fullscreen bool) {
		windowState.Maximized = fullscreen
		windowStore.Save(windowState)
	}
	closeModel := config.OnClosed
	config.OnClosed = func() {
		windowStore.Save(windowState)
		closeModel()
	}
	var canvas ui.Canvas
	err = app.RunApplication(func(application *app.Application) error {
		var err error
		window, err = application.OpenWindow(config, &canvas)
		if err != nil {
			return err
		}
		model.MaximizeWindow = func() { maximizeWindow(window) }
		model.RestoreWindow = func() { restoreWindow(window) }
		path, restore := startupProject(os.Args[1:], recentStore.Load(), preferences.RestoreLastProject)
		if restore {
			model.RestorePath(path, windowState.Maximized)
		} else if path != "" {
			model.OpenPath(path)
		}
		model.CheckForUpdates(true)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func startupProject(args, recent []string, allowRestore bool) (path string, restore bool) {
	if len(args) > 0 {
		return args[0], false
	}
	if allowRestore && len(recent) > 0 {
		return recent[0], true
	}
	return "", false
}
