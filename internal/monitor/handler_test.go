package monitor

import (
	"testing"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

func TestDefaultHandler_MatchesRule(t *testing.T) {
	handler := &DefaultHandler{}

	tests := []struct {
		name    string
		event   ConnectionEvent
		rule    rules.Rule
		matches bool
	}{
		{
			name: "exact match",
			event: ConnectionEvent{
				AppPath:   "C:\\Program Files\\App\\app.exe",
				Protocol:  "tcp",
				Direction: "outbound",
				DstPort:   443,
			},
			rule: rules.Rule{
				Application: "C:\\Program Files\\App\\app.exe",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{443},
			},
			matches: true,
		},
		{
			name: "protocol mismatch",
			event: ConnectionEvent{
				AppPath:   "C:\\Program Files\\App\\app.exe",
				Protocol:  "udp",
				Direction: "outbound",
				DstPort:   443,
			},
			rule: rules.Rule{
				Application: "C:\\Program Files\\App\\app.exe",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{443},
			},
			matches: false,
		},
		{
			name: "port mismatch",
			event: ConnectionEvent{
				AppPath:   "C:\\Program Files\\App\\app.exe",
				Protocol:  "tcp",
				Direction: "outbound",
				DstPort:   80,
			},
			rule: rules.Rule{
				Application: "C:\\Program Files\\App\\app.exe",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{443},
			},
			matches: false,
		},
		{
			name: "wildcard app",
			event: ConnectionEvent{
				AppPath:   "C:\\Program Files\\App\\app.exe",
				Protocol:  "tcp",
				Direction: "outbound",
				DstPort:   443,
			},
			rule: rules.Rule{
				Application: "",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{443},
			},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.matchesRule(tt.event, tt.rule)
			if result != tt.matches {
				t.Errorf("matchesRule() = %v, want %v", result, tt.matches)
			}
		})
	}
}

func TestSanitizeForRuleName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "C:\\Program Files\\Chrome\\chrome.exe",
			expected: "chrome",
		},
		{
			path:     "/usr/bin/firefox",
			expected: "firefox",
		},
		{
			path:     "C:\\App (v2)\\my-app.exe",
			expected: "my_app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := sanitizeForRuleName(tt.path)
			if result != tt.expected {
				t.Errorf("sanitizeForRuleName(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}
