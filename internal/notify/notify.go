package notify

import (
	"fmt"
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

// Show displays a notification with a title and message, returning the user's choice.
// Returns "yes" if user allows, "no" if user denies.
func Show(title, message string) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return showWindows(title, message)
	case "linux":
		return showLinux(title, message)
	default:
		return "no", fmt.Errorf("notifications not supported on %s", runtime.GOOS)
	}
}
