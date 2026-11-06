//go:build !windows

package main

import "github.com/vivaui/viva/app"

func maximizeWindow(window *app.Window) {
	_ = window.SetFullscreen(true)
}

func restoreWindow(window *app.Window) {
	_ = window.SetFullscreen(false)
}
