package platform

import (
	"fmt"
	"runtime"

	lin "github.com/vhPedroGitHub/firewall/internal/platform/linux"
	win "github.com/vhPedroGitHub/firewall/internal/platform/windows"
	"github.com/vhPedroGitHub/firewall/internal/rules"
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
