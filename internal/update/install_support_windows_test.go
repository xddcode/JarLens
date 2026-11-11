package update

import (
	"path/filepath"
	"testing"
)

func TestDetectWindowsInstall(t *testing.T) {
	custom := filepath.Join("C:\\", "Custom", "JarLens")
	programFiles := filepath.Join("C:\\", "Program Files", "JarLens")
	legacy := filepath.Join("C:\\", "Users", "test", "AppData", "Local", "Programs", "JarLens")
	tests := []struct {
		name       string
		executable string
		configure  func(*windowsInstallEnv)
		want       InstallReason
	}{
		{name: "registered per-user install", executable: filepath.Join(custom, "jarlens.exe"), configure: func(env *windowsInstallEnv) { env.userDir = custom }, want: InstallSupported},
		{name: "registered custom program files install", executable: filepath.Join(programFiles, "jarlens.exe"), configure: func(env *windowsInstallEnv) { env.userDir = programFiles }, want: InstallSupported},
		{name: "portable", executable: filepath.Join("D:\\", "Downloads", "JarLens", "jarlens.exe"), want: InstallNotUserInstall},
		{name: "system install", executable: filepath.Join(custom, "jarlens.exe"), configure: func(env *windowsInstallEnv) { env.machineDir = custom }, want: InstallSystemInstall},
		{name: "program files install", executable: filepath.Join(programFiles, "jarlens.exe"), want: InstallSystemInstall},
		{name: "legacy per-user install", executable: filepath.Join(legacy, "jarlens.exe"), want: InstallSupported},
		{name: "read-only install", executable: filepath.Join(custom, "jarlens.exe"), configure: func(env *windowsInstallEnv) {
			env.userDir = custom
			env.writable = func(string) bool { return false }
		}, want: InstallNotWritable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := windowsInstallEnv{
				localAppData: filepath.Join("C:\\", "Users", "test", "AppData", "Local"),
				programFiles: []string{filepath.Join("C:\\", "Program Files")},
				exists:       func(string) bool { return true },
				writable:     func(string) bool { return true },
			}
			if test.configure != nil {
				test.configure(&env)
			}
			if got := detectWindowsInstall(test.executable, env); got != test.want {
				t.Fatalf("reason = %v, want %v", got, test.want)
			}
		})
	}
}
