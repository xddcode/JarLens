package update

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const windowsUninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\io.github.xddcode.jarlens`

type windowsInstallEnv struct {
	localAppData string
	programFiles []string
	userDir      string
	machineDir   string
	writable     func(string) bool
	exists       func(string) bool
}

func detectInstallReason() InstallReason {
	executable, err := os.Executable()
	if err != nil {
		return InstallNotUserInstall
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return InstallNotUserInstall
	}
	userDir, machineDir := windowsInstallDirs()
	var programFiles []string
	for _, name := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
		if directory := os.Getenv(name); directory != "" {
			programFiles = append(programFiles, directory)
		}
	}
	return detectWindowsInstall(executable, windowsInstallEnv{
		localAppData: os.Getenv("LOCALAPPDATA"), programFiles: programFiles,
		userDir: userDir, machineDir: machineDir, writable: installDirWritable, exists: installPathExists,
	})
}

func detectWindowsInstall(executable string, env windowsInstallEnv) InstallReason {
	if runtime.GOOS != "windows" {
		return InstallNotUserInstall
	}
	directory := filepath.Dir(executable)
	registered := env.userDir != "" && strings.EqualFold(filepath.Clean(directory), filepath.Clean(env.userDir))
	if env.machineDir != "" && strings.EqualFold(filepath.Clean(directory), filepath.Clean(env.machineDir)) {
		return InstallSystemInstall
	}
	for _, root := range env.programFiles {
		if !registered && root != "" && installPathInside(executable, root) {
			return InstallSystemInstall
		}
	}
	programs := filepath.Join(env.localAppData, "Programs")
	legacyDefault := env.userDir == "" && env.localAppData != "" && strings.EqualFold(filepath.Dir(directory), programs) && strings.EqualFold(filepath.Base(directory), "JarLens")
	if (!registered && !legacyDefault) || !env.exists(filepath.Join(directory, "Uninstall.exe")) {
		return InstallNotUserInstall
	}
	if !env.writable(directory) {
		return InstallNotWritable
	}
	return InstallSupported
}

func windowsInstallDirs() (user, machine string) {
	read := func(root registry.Key) string {
		key, err := registry.OpenKey(root, windowsUninstallKey, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			return ""
		}
		defer key.Close()
		directory, _, err := key.GetStringValue("InstallLocation")
		if err != nil {
			return ""
		}
		return directory
	}
	return read(registry.CURRENT_USER), read(registry.LOCAL_MACHINE)
}

func installPathInside(path, directory string) bool {
	relative, err := filepath.Rel(directory, path)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func installPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func installDirWritable(directory string) bool {
	file, err := os.CreateTemp(directory, ".jarlens-update-write-*")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	return os.Remove(name) == nil
}
