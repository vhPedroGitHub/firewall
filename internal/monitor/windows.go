//go:build windows
// +build windows

package monitor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// WindowsMonitor monitors network connections on Windows.
// This is a simplified implementation that polls netstat for new connections.
// A production implementation would use Windows Filtering Platform (WFP) APIs via CGO.
type WindowsMonitor struct {
	mu       sync.Mutex
	running  bool
	cancel   context.CancelFunc
	interval time.Duration
	seen     map[string]bool // Track seen connections to avoid duplicates
}

// NewWindowsMonitor creates a new Windows connection monitor.
func NewWindowsMonitor() *WindowsMonitor {
	return &WindowsMonitor{
		interval: 2 * time.Second,
		seen:     make(map[string]bool),
	}
}

// Start begins monitoring network connections.
func (m *WindowsMonitor) Start() (<-chan ConnectionEvent, error) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil, fmt.Errorf("monitor already running")
	}
	m.running = true
	m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	events := make(chan ConnectionEvent, 10)

	go m.monitorLoop(ctx, events)

	return events, nil
}

// Stop gracefully shuts down the monitor.
func (m *WindowsMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("monitor not running")
	}

	if m.cancel != nil {
		m.cancel()
	}

	m.running = false
	return nil
}

// monitorLoop polls for new connections.
func (m *WindowsMonitor) monitorLoop(ctx context.Context, events chan<- ConnectionEvent) {
	defer close(events)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkConnections(events)
		}
	}
}

// checkConnections polls netstat for active connections.
func (m *WindowsMonitor) checkConnections(events chan<- ConnectionEvent) {
	// Use netstat to get active connections with process names
	cmd := exec.Command("netstat", "-ano")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if event := m.parseNetstatLine(line); event != nil {
			// Create a unique key for this connection
			key := fmt.Sprintf("%s|%s|%s:%d|%s:%d",
				event.AppPath,
				event.Protocol,
				event.SrcAddr,
				event.SrcPort,
				event.DstAddr,
				event.DstPort,
			)

			m.mu.Lock()
			if !m.seen[key] {
				m.seen[key] = true
				m.mu.Unlock()

				// Send new connection event
				select {
				case events <- *event:
				default:
					// Channel full, skip this event
				}
			} else {
				m.mu.Unlock()
			}
		}
	}
}

// parseNetstatLine parses a netstat output line into a ConnectionEvent.
func (m *WindowsMonitor) parseNetstatLine(line string) *ConnectionEvent {
	// Example netstat line:
	// TCP    192.168.1.100:50123    93.184.216.34:443    ESTABLISHED    1234
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil
	}

	protocol := strings.ToLower(fields[0])
	if protocol != "tcp" && protocol != "udp" {
		return nil
	}

	// Parse local address
	localParts := strings.Split(fields[1], ":")
	if len(localParts) != 2 {
		return nil
	}
	srcAddr := localParts[0]
	srcPort, err := strconv.Atoi(localParts[1])
	if err != nil {
		return nil
	}

	// Parse foreign address
	foreignParts := strings.Split(fields[2], ":")
	if len(foreignParts) != 2 {
		return nil
	}
	dstAddr := foreignParts[0]
	dstPort, err := strconv.Atoi(foreignParts[1])
	if err != nil {
		return nil
	}

	// Get state (field 3 for TCP, not present for UDP)
	state := "N/A"
	pidField := 3
	if protocol == "tcp" && len(fields) >= 5 {
		state = fields[3]
		pidField = 4
	}

	// Determine direction (simplified - outbound if destination is not local)
	direction := "outbound"
	if !strings.HasPrefix(dstAddr, "192.168.") && !strings.HasPrefix(dstAddr, "10.") && !strings.HasPrefix(dstAddr, "172.") {
		direction = "outbound"
	} else if !strings.HasPrefix(srcAddr, "192.168.") && !strings.HasPrefix(srcAddr, "10.") && !strings.HasPrefix(srcAddr, "172.") {
		direction = "inbound"
	}

	// Get PID and resolve to process name
	var pid string
	var appPath string
	if len(fields) > pidField {
		pid = fields[pidField]
		appPath = m.getProcessPath(pid)
	}

	return &ConnectionEvent{
		AppPath:   appPath,
		PID:       pid,
		Protocol:  protocol,
		Direction: direction,
		SrcAddr:   srcAddr,
		SrcPort:   srcPort,
		DstAddr:   dstAddr,
		DstPort:   dstPort,
		State:     state,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// getProcessPath resolves a PID to the full executable path.
func (m *WindowsMonitor) getProcessPath(pid string) string {
	// Use WMIC to get process path
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "get", "ExecutablePath", "/value")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("PID:%s", pid)
	}

	// Parse output: "ExecutablePath=C:\Path\To\App.exe"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ExecutablePath=") {
			path := strings.TrimPrefix(line, "ExecutablePath=")
			path = strings.TrimSpace(path)
			if path != "" {
				return path
			}
		}
	}

	return fmt.Sprintf("PID:%s", pid)
}
