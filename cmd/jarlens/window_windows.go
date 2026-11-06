package main

import (
	"github.com/vivaui/viva/app"
	"golang.org/x/sys/windows"
)

var showWindow = windows.NewLazySystemDLL("user32.dll").NewProc("ShowWindow")

func maximizeWindow(_ *app.Window) {
	window := windows.GetForegroundWindow()
	if window != 0 {
		showWindow.Call(uintptr(window), uintptr(windows.SW_SHOWMAXIMIZED))
	}
}

func restoreWindow(_ *app.Window) {
	window := windows.GetForegroundWindow()
	if window != 0 {
		showWindow.Call(uintptr(window), uintptr(windows.SW_RESTORE))
	}
}
