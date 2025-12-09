//go:build linux
// +build linux

package notify

import (
	"fmt"
	"os/exec"
)

// promptLinux uses zenity for interactive dialog on Linux.
func promptLinux(app string) (bool, error) {
	// Try zenity for interactive dialog
	cmd := exec.Command("zenity", "--question", fmt.Sprintf("--text=Allow %s to connect?", app), "--title=Firewall")
	err := cmd.Run()
	if err == nil {
		return true, nil // User clicked Yes
	}

	// Exit code 1 means No, other codes are errors
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, fmt.Errorf("notification failed: %w", err)
}

// showLinux uses zenity for notifications on Linux.
func showLinux(title, message string) (string, error) {
	cmd := exec.Command("zenity", "--question", fmt.Sprintf("--text=%s", message), fmt.Sprintf("--title=%s", title))
	err := cmd.Run()
	if err == nil {
		return "yes", nil
	}

	// Exit code 1 means No, other codes are errors
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return "no", nil
	}

	return "no", fmt.Errorf("notification failed: %w", err)
}

// Windows stubs for Linux builds
func promptWindows(app string) (bool, error) {
	return false, fmt.Errorf("Windows prompts not supported on Linux")
}

func showWindows(title, message string) (string, error) {
	return "no", fmt.Errorf("Windows prompts not supported on Linux")
}
