//go:build !windows && !darwin && !linux

package workspace

import "errors"

func SetLaunchAtLogin(bool) error {
	return errors.New("automatic login startup is not supported on this platform")
}
