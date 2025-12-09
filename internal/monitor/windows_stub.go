//go:build !windows
// +build !windows

package monitor

import "fmt"

// WindowsMonitor stub for non-Windows platforms.
type WindowsMonitor struct{}

func NewWindowsMonitor() *WindowsMonitor {
	return &WindowsMonitor{}
}

func (m *WindowsMonitor) Start() (<-chan ConnectionEvent, error) {
	return nil, fmt.Errorf("windows monitor not available on this platform")
}

func (m *WindowsMonitor) Stop() error {
	return fmt.Errorf("windows monitor not available on this platform")
}
