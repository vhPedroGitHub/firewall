package monitor

import (
	"fmt"
	"runtime"
)

// New creates a platform-specific monitor.
func New() (Monitor, error) {
	switch runtime.GOOS {
	case "windows":
		return NewWindowsMonitor(), nil
	case "linux":
		return NewLinuxMonitor(), nil
	default:
		return nil, fmt.Errorf("monitoring not supported on %s", runtime.GOOS)
	}
}
