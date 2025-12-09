//go:build !windows

package windows

import (
	"fmt"

	"firewall/internal/rules"
)

func ApplyRule(r rules.Rule) error {
	_ = r
	return fmt.Errorf("windows adapter not available on this platform")
}
