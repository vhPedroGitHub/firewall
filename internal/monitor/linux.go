//go:build linux
// +build linux

package monitor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LinuxMonitor monitors network connections on Linux.
// This is a simplified implementation that reads /proc/net/tcp and /proc/net/udp.
// A production implementation would use netfilter/nfqueue for real-time monitoring.
type LinuxMonitor struct {
	mu       sync.Mutex
	running  bool
	cancel   context.CancelFunc
	interval time.Duration
	seen     map[string]bool
}

// NewLinuxMonitor creates a new Linux connection monitor.
func NewLinuxMonitor() *LinuxMonitor {
	return &LinuxMonitor{
		interval: 2 * time.Second,
		seen:     make(map[string]bool),
	}
}

// Start begins monitoring network connections.
func (m *LinuxMonitor) Start() (<-chan ConnectionEvent, error) {
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
func (m *LinuxMonitor) Stop() error {
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
func (m *LinuxMonitor) monitorLoop(ctx context.Context, events chan<- ConnectionEvent) {
	defer close(events)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkConnections(events, "tcp", "/proc/net/tcp")
			m.checkConnections(events, "udp", "/proc/net/udp")
		}
	}
}

// checkConnections reads /proc/net files for active connections.
func (m *LinuxMonitor) checkConnections(events chan<- ConnectionEvent, protocol, procFile string) {
	file, err := os.Open(procFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header line
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if event := m.parseProcNetLine(line, protocol); event != nil {
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

				select {
				case events <- *event:
				default:
				}
			} else {
				m.mu.Unlock()
			}
		}
	}
}

// parseProcNetLine parses a /proc/net/tcp or /proc/net/udp line.
func (m *LinuxMonitor) parseProcNetLine(line, protocol string) *ConnectionEvent {
	// Example line from /proc/net/tcp:
	//   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 ...
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return nil
	}

	// Parse local address (field 1)
	localAddr, localPort := parseHexAddress(fields[1])
	if localAddr == "" {
		return nil
	}

	// Parse remote address (field 2)
	remoteAddr, remotePort := parseHexAddress(fields[2])
	if remoteAddr == "" {
		return nil
	}

	// Get state (field 3) - hex value
	stateHex := fields[3]
	state := tcpStateToString(stateHex)

	// Get inode (field 9)
	inode := fields[9]

	// Resolve process from inode and get PID
	appPath, pid := m.getProcessByInodeWithPID(inode)

	// Determine direction
	direction := "outbound"
	if isLocalAddress(localAddr) && !isLocalAddress(remoteAddr) {
		direction = "outbound"
	} else if !isLocalAddress(localAddr) && isLocalAddress(remoteAddr) {
		direction = "inbound"
	}

	return &ConnectionEvent{
		AppPath:   appPath,
		PID:       pid,
		Protocol:  protocol,
		Direction: direction,
		SrcAddr:   localAddr,
		SrcPort:   localPort,
		DstAddr:   remoteAddr,
		DstPort:   remotePort,
		State:     state,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// parseHexAddress converts hex address format to IP:port.
func parseHexAddress(hexAddr string) (string, int) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return "", 0
	}

	// Convert hex IP to decimal
	ipHex := parts[0]
	if len(ipHex) != 8 {
		return "", 0
	}

	// IP is in little-endian format
	ip := fmt.Sprintf("%d.%d.%d.%d",
		hexToInt(ipHex[6:8]),
		hexToInt(ipHex[4:6]),
		hexToInt(ipHex[2:4]),
		hexToInt(ipHex[0:2]),
	)

	// Convert hex port to decimal
	port := hexToInt(parts[1])

	return ip, port
}

// hexToInt converts a hex string to int.
func hexToInt(hex string) int {
	val, _ := strconv.ParseInt(hex, 16, 64)
	return int(val)
}

// isLocalAddress checks if an IP is a local/private address.
func isLocalAddress(ip string) bool {
	return strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "127.") ||
		ip == "0.0.0.0"
}

// getProcessByInode finds the process that owns a socket inode.
func (m *LinuxMonitor) getProcessByInode(inode string) string {
	appPath, _ := m.getProcessByInodeWithPID(inode)
	return appPath
}

// getProcessByInodeWithPID finds the process and PID that owns a socket inode.
func (m *LinuxMonitor) getProcessByInodeWithPID(inode string) (string, string) {
	// Search /proc/*/fd/* for the inode
	procDirs, err := filepath.Glob("/proc/[0-9]*/fd/*")
	if err != nil {
		return fmt.Sprintf("inode:%s", inode), ""
	}

	searchStr := fmt.Sprintf("socket:[%s]", inode)
	for _, fdPath := range procDirs {
		link, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}

		if link == searchStr {
			// Found the process - get its exe path and PID
			pidStr := strings.Split(fdPath, "/")[2]
			exePath := fmt.Sprintf("/proc/%s/exe", pidStr)
			realPath, err := os.Readlink(exePath)
			if err == nil {
				return realPath, pidStr
			}
			return fmt.Sprintf("PID:%s", pidStr), pidStr
		}
	}

	return fmt.Sprintf("inode:%s", inode), ""
}

// tcpStateToString converts hex TCP state to readable string.
func tcpStateToString(stateHex string) string {
	state := hexToInt(stateHex)
	states := map[int]string{
		0x01: "ESTABLISHED",
		0x02: "SYN_SENT",
		0x03: "SYN_RECV",
		0x04: "FIN_WAIT1",
		0x05: "FIN_WAIT2",
		0x06: "TIME_WAIT",
		0x07: "CLOSE",
		0x08: "CLOSE_WAIT",
		0x09: "LAST_ACK",
		0x0A: "LISTEN",
		0x0B: "CLOSING",
	}
	if s, ok := states[state]; ok {
		return s
	}
	return "UNKNOWN"
}

// getProcessByInode finds the process that owns a socket inode (deprecated).
func (m *LinuxMonitor) getProcessByInodeOld(inode string) string {
	// Search /proc/*/fd/* for the inode
	procDirs, err := filepath.Glob("/proc/[0-9]*/fd/*")
	if err != nil {
		return fmt.Sprintf("inode:%s", inode)
	}

	searchStr := fmt.Sprintf("socket:[%s]", inode)
	for _, fdPath := range procDirs {
		link, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}

		if link == searchStr {
			// Found the process - get its exe path
			pidStr := strings.Split(fdPath, "/")[2]
			exePath := fmt.Sprintf("/proc/%s/exe", pidStr)
			realPath, err := os.Readlink(exePath)
			if err == nil {
				return realPath
			}
			return fmt.Sprintf("PID:%s", pidStr)
		}
	}

	return fmt.Sprintf("inode:%s", inode)
}
