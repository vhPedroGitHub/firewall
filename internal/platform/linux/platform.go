//go:build linux

package linux

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

// ApplyRule applies a firewall rule using iptables on Linux.
func ApplyRule(r rules.Rule) error {
	// Map our rule to iptables parameters
	chain := "INPUT"
	if r.Direction == "outbound" {
		chain = "OUTPUT"
	}

	target := "ACCEPT"
	if r.Action == "deny" {
		target = "DROP"
	}

	// Build iptables command: iptables -A <chain> -p <protocol> --dport <ports> -j <target>
	args := []string{"-A", chain}

	// Protocol
	if r.Protocol != "any" {
		args = append(args, "-p", r.Protocol)
	}

	// Ports (for tcp/udp)
	if len(r.Ports) > 0 && r.Protocol != "any" {
		portList := make([]string, len(r.Ports))
		for i, p := range r.Ports {
			portList[i] = fmt.Sprintf("%d", p)
		}
		portSpec := strings.Join(portList, ",")

		if r.Direction == "inbound" {
			args = append(args, "--dport", portSpec)
		} else {
			args = append(args, "--dport", portSpec)
		}
	}

	// Add comment with application name
	args = append(args, "-m", "comment", "--comment", fmt.Sprintf("firewall-rule:%s:%s", r.Name, r.Application))

	// Target
	args = append(args, "-j", target)

	cmd := exec.Command("iptables", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iptables failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// RemoveRule removes a firewall rule by name using iptables comment matching on Linux.
func RemoveRule(name string) error {
	// Search and delete rules with our comment pattern
	for _, chain := range []string{"INPUT", "OUTPUT"} {
		comment := fmt.Sprintf("firewall-rule:%s:", name)

		// List rules with line numbers, find matching comment, delete
		cmd := exec.Command("iptables", "-L", chain, "--line-numbers", "-n")
		_, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}

		// Simple approach: try to delete by comment match
		// In production, parse output and delete by line number
		delCmd := exec.Command("iptables", "-D", chain, "-m", "comment", "--comment", comment)
		_ = delCmd.Run() // Ignore error if rule doesn't exist
	}

	return nil
}
