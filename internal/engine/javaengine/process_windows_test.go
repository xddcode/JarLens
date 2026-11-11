//go:build windows

package javaengine

import (
	"os/exec"
	"testing"
)

func TestConfigureProcessHidesConsoleWindow(t *testing.T) {
	command := exec.Command("java")
	configureProcess(command)
	if command.SysProcAttr == nil || !command.SysProcAttr.HideWindow {
		t.Fatal("Java process would show a console window")
	}
}
