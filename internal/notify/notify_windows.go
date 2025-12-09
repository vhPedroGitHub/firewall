//go:build windows
// +build windows

package notify

import (
	"fmt"
	"os/exec"
	"syscall"
)

// promptWindows uses PowerShell MessageBox for notification on Windows.
func promptWindows(app string) (bool, error) {
	// Use PowerShell to show a simple message box
	script := fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('Allow %s to connect?', 'Firewall', 'YesNo')`, app)
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("notification failed: %w", err)
	}

	// Check if user clicked "Yes"
	result := string(output)
	if len(result) > 0 && result[0] == 'Y' {
		return true, nil
	}
	return false, nil
}

// showWindows uses PowerShell MessageBox for notifications on Windows.
func showWindows(title, message string) (string, error) {
	script := fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('%s', '%s', 'YesNo')`, message, title)
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "no", fmt.Errorf("notification failed: %w", err)
	}

	result := string(output)
	if len(result) > 0 && (result[0] == 'Y' || result[0] == 'y') {
		return "yes", nil
	}
	return "no", nil
}

// Linux stubs for Windows builds
func promptLinux(app string) (bool, error) {
	return false, fmt.Errorf("Linux prompts not supported on Windows")
}

func showLinux(title, message string) (string, error) {
	return "no", fmt.Errorf("Linux prompts not supported on Windows")
}
