//go:build !linux

package linux

import (
	"fmt"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

func ApplyRule(r rules.Rule) error {
	_ = r
	return fmt.Errorf("linux adapter not available on this platform")
}
