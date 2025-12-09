//go:build windows

package windows

import (
	"fmt"
	"os/exec"
	"strings"

	"firewall/internal/rules"
)

// ApplyRule applies a firewall rule using netsh advfirewall on Windows.
func ApplyRule(r rules.Rule) error {
	// Map our rule to netsh parameters
	dir := "in"
	if r.Direction == "outbound" {
		dir = "out"
	}

	action := "allow"
	if r.Action == "deny" {
		action = "block"
	}

	protocol := r.Protocol
	if protocol == "any" {
		protocol = "any"
	}

	// Build netsh command: netsh advfirewall firewall add rule ...
	args := []string{
		"advfirewall", "firewall", "add", "rule",
		fmt.Sprintf("name=%s", r.Name),
		fmt.Sprintf("dir=%s", dir),
		fmt.Sprintf("action=%s", action),
		fmt.Sprintf("program=%s", r.Application),
		fmt.Sprintf("protocol=%s", protocol),
	}

	// Add port specification if needed
	if len(r.Ports) > 0 && protocol != "any" {
		portList := make([]string, len(r.Ports))
		for i, p := range r.Ports {
			portList[i] = fmt.Sprintf("%d", p)
		}
		portSpec := strings.Join(portList, ",")
		if dir == "in" {
			args = append(args, fmt.Sprintf("localport=%s", portSpec))
		} else {
			args = append(args, fmt.Sprintf("remoteport=%s", portSpec))
		}
	}

	cmd := exec.Command("netsh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// RemoveRule removes a firewall rule by name using netsh on Windows.
func RemoveRule(name string) error {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", fmt.Sprintf("name=%s", name))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh delete failed: %w (output: %s)", err, string(output))
	}
	return nil
}
