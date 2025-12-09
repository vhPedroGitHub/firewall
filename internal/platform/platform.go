package platform

import (
	"fmt"
	"runtime"

	lin "firewall/internal/platform/linux"
	win "firewall/internal/platform/windows"
	"firewall/internal/rules"
)

// ApplyRule dispatches to the OS-specific adapter.
func ApplyRule(r rules.Rule) error {
	switch runtime.GOOS {
	case "windows":
		return win.ApplyRule(r)
	case "linux":
		return lin.ApplyRule(r)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
