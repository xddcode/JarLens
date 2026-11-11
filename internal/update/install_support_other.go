//go:build !windows

package update

func detectInstallReason() InstallReason {
	return InstallNotUserInstall
}
