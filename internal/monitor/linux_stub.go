//go:build !linux
// +build !linux

package monitor

import "fmt"

// LinuxMonitor stub for non-Linux platforms.
type LinuxMonitor struct{}

func NewLinuxMonitor() *LinuxMonitor {
	return &LinuxMonitor{}
}

func (m *LinuxMonitor) Start() (<-chan ConnectionEvent, error) {
	return nil, fmt.Errorf("linux monitor not available on this platform")
}

func (m *LinuxMonitor) Stop() error {
	return fmt.Errorf("linux monitor not available on this platform")
}
