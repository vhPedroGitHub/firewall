//go:build !windows && !linux
// +build !windows,!linux

package notify

import "fmt"

func promptWindows(app string) (bool, error) {
	return false, fmt.Errorf("not supported on this platform")
}

func showWindows(title, message string) (string, error) {
	return "no", fmt.Errorf("not supported on this platform")
}

func promptLinux(app string) (bool, error) {
	return false, fmt.Errorf("not supported on this platform")
}

func showLinux(title, message string) (string, error) {
	return "no", fmt.Errorf("not supported on this platform")
}
