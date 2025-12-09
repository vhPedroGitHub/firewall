package rules

import "fmt"

// Rule represents a single firewall rule configuration.
type Rule struct {
	Name        string
	Application string
	Action      string // allow or deny
	Protocol    string // tcp, udp, any
	Ports       []int
	Direction   string // inbound or outbound
}

// Validate performs basic rule validation; expand with richer checks later.
func Validate(r Rule) error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Application == "" {
		return fmt.Errorf("application is required")
	}

	switch r.Action {
	case "allow", "deny":
	default:
		return fmt.Errorf("invalid action: %s", r.Action)
	}

	switch r.Protocol {
	case "tcp", "udp", "any":
	default:
		return fmt.Errorf("invalid protocol: %s", r.Protocol)
	}

	switch r.Direction {
	case "inbound", "outbound":
	default:
		return fmt.Errorf("invalid direction: %s", r.Direction)
	}

	if r.Protocol != "any" {
		if len(r.Ports) == 0 {
			return fmt.Errorf("ports required for protocol %s", r.Protocol)
		}
		for _, p := range r.Ports {
			if p <= 0 || p > 65535 {
				return fmt.Errorf("invalid port: %d", p)
			}
		}
	}

	return nil
}
