package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

func SetLaunchAtLogin(enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "autostart", "jarlens.desktop")
	if !enabled {
		return removeLaunchItem(path)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return err
	}
	quoted := `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`).Replace(executable) + `"`
	content := `[Desktop Entry]
Type=Application
Version=1.0
Name=JarLens
Comment=Java decompiler and code explorer
Exec=` + quoted + `
Terminal=false
StartupNotify=false
X-GNOME-Autostart-enabled=true
`
	return writeLaunchItem(path, []byte(content))
}
