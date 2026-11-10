package javaengine

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func findJava(executablePath string) string {
	if override := os.Getenv("JARLENS_JAVA"); override != "" {
		return override
	}
	if executablePath == "" {
		executablePath, _ = os.Executable()
	}
	name := "java"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	directory := filepath.Dir(executablePath)
	for _, candidate := range []string{
		filepath.Join(directory, "jre", "bin", name),
		filepath.Join(directory, "..", "Resources", "jre", "bin", name),
	} {
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	java, _ := exec.LookPath("java")
	return java
}
