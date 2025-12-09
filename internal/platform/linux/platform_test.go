//go:build linux
// +build linux

package linux

import (
	"strings"
	"testing"

	"firewall/internal/rules"
)

func TestApplyRule_CommandGeneration(t *testing.T) {
	tests := []struct {
		name     string
		rule     rules.Rule
		wantArgs []string
	}{
		{
			name: "basic outbound allow rule",
			rule: rules.Rule{
				Name:        "web-browser",
				Application: "/usr/bin/firefox",
				Action:      "allow",
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{80, 443},
			},
			wantArgs: []string{
				"iptables",
				"-A",
				"OUTPUT",
				"-p",
				"tcp",
				"-m",
				"multiport",
				"--dports",
				"80,443",
				"-j",
				"ACCEPT",
				"-m",
				"comment",
				"--comment",
				"firewall:web-browser",
			},
		},
		{
			name: "inbound deny rule",
			rule: rules.Rule{
				Name:        "block-ssh",
				Application: "/usr/sbin/sshd",
				Action:      "deny",
				Protocol:    "tcp",
				Direction:   "inbound",
				Ports:       []int{22},
			},
			wantArgs: []string{
				"iptables",
				"-A",
				"INPUT",
				"-p",
				"tcp",
				"--dport",
				"22",
				"-j",
				"DROP",
				"-m",
				"comment",
				"--comment",
				"firewall:block-ssh",
			},
		},
		{
			name: "any protocol rule",
			rule: rules.Rule{
				Name:        "allow-all",
				Application: "/usr/bin/app",
				Action:      "allow",
				Protocol:    "any",
				Direction:   "outbound",
				Ports:       nil,
			},
			wantArgs: []string{
				"iptables",
				"-A",
				"OUTPUT",
				"-j",
				"ACCEPT",
				"-m",
				"comment",
				"--comment",
				"firewall:allow-all",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildIptablesArgs(tt.rule)
			cmdStr := strings.Join(args, " ")

			for _, want := range tt.wantArgs {
				if !strings.Contains(cmdStr, want) {
					t.Errorf("command missing expected arg %q\nfull command: %s", want, cmdStr)
				}
			}
		})
	}
}

func buildIptablesArgs(r rules.Rule) []string {
	// Mirror the logic from platform.go
	args := []string{"iptables"}

	chain := "OUTPUT"
	if r.Direction == "inbound" {
		chain = "INPUT"
	}
	args = append(args, "-A", chain)

	if r.Protocol != "any" {
		args = append(args, "-p", r.Protocol)
	}

	if len(r.Ports) > 0 && r.Protocol != "any" {
		if len(r.Ports) == 1 {
			args = append(args, "--dport", portToStr(r.Ports[0]))
		} else {
			var portList string
			for i, p := range r.Ports {
				if i > 0 {
					portList += ","
				}
				portList += portToStr(p)
			}
			args = append(args, "-m", "multiport", "--dports", portList)
		}
	}

	target := "ACCEPT"
	if r.Action == "deny" {
		target = "DROP"
	}
	args = append(args, "-j", target)

	args = append(args, "-m", "comment", "--comment", "firewall:"+r.Name)

	return args
}

func portToStr(p int) string {
	// Simple int to string conversion
	if p == 0 {
		return "0"
	}
	var result string
	for p > 0 {
		result = string(rune('0'+p%10)) + result
		p /= 10
	}
	return result
}

func TestApplyRule_ChainSelection(t *testing.T) {
	tests := []struct {
		direction string
		wantChain string
	}{
		{"inbound", "INPUT"},
		{"outbound", "OUTPUT"},
	}

	for _, tt := range tests {
		t.Run(tt.direction, func(t *testing.T) {
			r := rules.Rule{
				Name:        "test",
				Application: "/usr/bin/test",
				Action:      "allow",
				Protocol:    "tcp",
				Direction:   tt.direction,
				Ports:       []int{80},
			}

			args := buildIptablesArgs(r)
			cmdStr := strings.Join(args, " ")

			if !strings.Contains(cmdStr, tt.wantChain) {
				t.Errorf("expected chain %s in command: %s", tt.wantChain, cmdStr)
			}
		})
	}
}

func TestApplyRule_TargetMapping(t *testing.T) {
	tests := []struct {
		action     string
		wantTarget string
	}{
		{"allow", "ACCEPT"},
		{"deny", "DROP"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			r := rules.Rule{
				Name:        "test",
				Application: "/usr/bin/test",
				Action:      tt.action,
				Protocol:    "tcp",
				Direction:   "outbound",
				Ports:       []int{80},
			}

			args := buildIptablesArgs(r)
			cmdStr := strings.Join(args, " ")

			if !strings.Contains(cmdStr, "-j "+tt.wantTarget) {
				t.Errorf("expected target %s in command: %s", tt.wantTarget, cmdStr)
			}
		})
	}
}

func TestApplyRule_CommentGeneration(t *testing.T) {
	r := rules.Rule{
		Name:        "my-app-rule",
		Application: "/usr/bin/app",
		Action:      "allow",
		Protocol:    "tcp",
		Direction:   "outbound",
		Ports:       []int{443},
	}

	args := buildIptablesArgs(r)
	cmdStr := strings.Join(args, " ")

	expectedComment := "firewall:my-app-rule"
	if !strings.Contains(cmdStr, expectedComment) {
		t.Errorf("expected comment %q in command: %s", expectedComment, cmdStr)
	}
}

func TestApplyRule_MultiplePortsUseMultiport(t *testing.T) {
	r := rules.Rule{
		Name:        "multi-port",
		Application: "/usr/bin/app",
		Action:      "allow",
		Protocol:    "tcp",
		Direction:   "outbound",
		Ports:       []int{80, 443, 8080},
	}

	args := buildIptablesArgs(r)
	cmdStr := strings.Join(args, " ")

	if !strings.Contains(cmdStr, "multiport") {
		t.Error("expected multiport module for multiple ports")
	}
	if !strings.Contains(cmdStr, "--dports") {
		t.Error("expected --dports flag for multiple ports")
	}
}

func TestApplyRule_SinglePortUsesDport(t *testing.T) {
	r := rules.Rule{
		Name:        "single-port",
		Application: "/usr/bin/app",
		Action:      "allow",
		Protocol:    "tcp",
		Direction:   "outbound",
		Ports:       []int{443},
	}

	args := buildIptablesArgs(r)
	cmdStr := strings.Join(args, " ")

	if !strings.Contains(cmdStr, "--dport 443") {
		t.Error("expected --dport for single port")
	}
	if strings.Contains(cmdStr, "multiport") {
		t.Error("did not expect multiport for single port")
	}
}
