package monitor

import (
	"github.com/vhPedroGitHub/firewall/internal/rules"
)

// ConnectionEvent represents a detected network connection attempt.
type ConnectionEvent struct {
	AppPath   string // Full path to the application making the connection
	PID       string // Process ID
	Protocol  string // tcp, udp, icmp
	Direction string // inbound, outbound
	SrcAddr   string // Source IP address
	SrcPort   int    // Source port
	DstAddr   string // Destination IP address
	DstPort   int    // Destination port
	State     string // Connection state (ESTABLISHED, LISTENING, TIME_WAIT, etc.)
	Timestamp string // Time when the connection was detected
}

// Decision represents the user's choice for a connection.
type Decision int

const (
	DecisionAllow Decision = iota
	DecisionDeny
	DecisionCancel
)

// Monitor defines the interface for connection monitoring.
type Monitor interface {
	// Start begins monitoring network connections.
	// Returns a channel for connection events and an error if startup fails.
	Start() (<-chan ConnectionEvent, error)

	// Stop gracefully shuts down the monitor.
	Stop() error
}

// Handler processes connection events and returns decisions.
type Handler interface {
	// HandleConnection is called when a new connection is detected.
	// It should check existing rules and prompt the user if needed.
	HandleConnection(event ConnectionEvent) (Decision, error)
}

// RuleChecker checks if a connection matches existing rules.
type RuleChecker interface {
	// CheckRule returns the matching rule's action (allow/deny) or nil if no match.
	CheckRule(event ConnectionEvent) *rules.Rule
}
