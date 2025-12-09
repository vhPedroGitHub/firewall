package notify

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Response represents a user's response to a notification prompt.
type Response struct {
	Allow bool
	Error error
}

// PromptUser displays a notification asking whether to allow or deny an application.
func PromptUser(app string) (bool, error) {
	switch runtime.GOOS {
	case "windows":
		return promptWindows(app)
	case "linux":
		return promptLinux(app)
	default:
		return false, fmt.Errorf("notifications not supported on %s", runtime.GOOS)
	}
}

// promptWindows uses msg.exe or PowerShell for basic notification on Windows.
func promptWindows(app string) (bool, error) {
	// Use PowerShell to show a simple message box
	script := fmt.Sprintf(`Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show('Allow %s to connect?', 'Firewall', 'YesNo')`, app)
	cmd := exec.Command("powershell", "-Command", script)
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

// promptLinux uses notify-send or zenity for notifications on Linux.
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
