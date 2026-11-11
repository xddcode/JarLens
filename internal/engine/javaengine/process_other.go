//go:build !windows

package javaengine

import "os/exec"

func configureProcess(*exec.Cmd) {}
