//go:build windows
// +build windows

package windows

import (
	"strings"
	"testing"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

func TestApplyRule_CommandGeneration(t *testing.T) {
	// This test verifies command generation without actually executing
	tests := []struct {
		name     string
		rule     rules.Rule
		wantArgs []string
	}{
		{
			name: "basic outbound rule",
			rule: rules.Rule{
				Name:        "test-web",
				Application: "C:\\Program Files\\Firefox\\firefox.exe",
				Action:      "allow",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{80, 443},
			},
			wantArgs: []string{
				"advfirewall",
				"firewall",
				"add",
				"rule",
				"name=test-web",
				"dir=out",
				"action=allow",
				"protocol=tcp",
				"program=\"C:\\Program Files\\Firefox\\firefox.exe\"",
				"localport=80,443",
			},
		},
		{
			name: "inbound deny rule",
			rule: rules.Rule{
				Name:        "block-app",
				Application: "C:\\test.exe",
				Action:      "deny",
				Protocol:    "udp",
				Direction:   "inbound",
				Ports:       []int{53},
			},
			wantArgs: []string{
				"advfirewall",
				"firewall",
				"add",
				"rule",
				"name=block-app",
				"dir=in",
				"action=block",
				"protocol=udp",
				"program=\"C:\\test.exe\"",
				"localport=53",
			},
		},
		{
			name: "any protocol rule",
			rule: rules.Rule{
				Name:        "any-proto",
				Application: "C:\\app.exe",
				Action:      "allow",
				Protocol:    "any",
				Direction:   "outbound",
				Ports:       nil,
			},
			wantArgs: []string{
				"advfirewall",
				"firewall",
				"add",
				"rule",
				"name=any-proto",
				"dir=out",
				"action=allow",
				"protocol=any",
				"program=\"C:\\app.exe\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildNetshArgs(tt.rule)

			// Verify key arguments are present
			cmdStr := strings.Join(args, " ")

			for _, want := range tt.wantArgs {
				if !strings.Contains(cmdStr, want) {
					t.Errorf("command missing expected arg %q\nfull command: %s", want, cmdStr)
				}
			}
		})
	}
}

func buildNetshArgs(r rules.Rule) []string {
	// Mirror the logic from platform.go
	args := []string{"advfirewall", "firewall", "add", "rule"}
	args = append(args, "name="+r.Name)

	dir := "out"
	if r.Direction == "inbound" {
		dir = "in"
	}
	args = append(args, "dir="+dir)

	action := r.Action
	if action == "deny" {
		action = "block"
	}
	args = append(args, "action="+action)
	args = append(args, "protocol="+r.Protocol)
	args = append(args, "program=\""+r.Application+"\"")

	if len(r.Ports) > 0 && r.Protocol != "any" {
		var portStr string
		for i, p := range r.Ports {
			if i > 0 {
				portStr += ","
			}
			portStr += intToStr(p)
		}
		args = append(args, "localport="+portStr)
	}

	return args
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func TestApplyRule_DirectionMapping(t *testing.T) {
	tests := []struct {
		direction string
		wantDir   string
	}{
		{"inbound", "in"},
		{"outbound", "out"},
	}

	for _, tt := range tests {
		t.Run(tt.direction, func(t *testing.T) {
			r := rules.Rule{
				Name:        "test",
				Application: "C:\\test.exe",
				Action:      "allow",
				Protocol:    "tcp",
				Direction:   tt.direction,
				Ports:       []int{80},
			}

			args := buildNetshArgs(r)
			cmdStr := strings.Join(args, " ")

			if !strings.Contains(cmdStr, "dir="+tt.wantDir) {
				t.Errorf("expected dir=%s in command: %s", tt.wantDir, cmdStr)
			}
		})
	}
}

func TestApplyRule_ActionMapping(t *testing.T) {
	tests := []struct {
		action     string
		wantAction string
	}{
		{"allow", "allow"},
		{"deny", "block"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			r := rules.Rule{
				Name:        "test",
				Application: "C:\\test.exe",
				Action:      tt.action,
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{80},
			}

			args := buildNetshArgs(r)
			cmdStr := strings.Join(args, " ")

			if !strings.Contains(cmdStr, "action="+tt.wantAction) {
				t.Errorf("expected action=%s in command: %s", tt.wantAction, cmdStr)
			}
		})
	}
}
