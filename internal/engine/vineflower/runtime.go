package vineflower

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const javaOverrideEnv = "JARLENS_JAVA"

func findJava(executablePath string) string {
	if override := os.Getenv(javaOverrideEnv); override != "" {
		return override
	}
	if executablePath == "" {
		executablePath, _ = os.Executable()
	}
	for _, candidate := range javaCandidates(executablePath) {
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	java, _ := exec.LookPath("java")
	return java
}

func javaCandidates(executablePath string) []string {
	name := "java"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	directory := filepath.Dir(executablePath)
	return []string{
		filepath.Join(directory, "jre", "bin", name),
		filepath.Join(directory, "..", "Resources", "jre", "bin", name),
	}
}
