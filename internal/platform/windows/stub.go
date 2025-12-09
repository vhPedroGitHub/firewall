//go:build !windows

package windows

import (
	"fmt"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

func ApplyRule(r rules.Rule) error {
	_ = r
	return fmt.Errorf("windows adapter not available on this platform")
}
