//go:build !linux

package linux

import (
	"fmt"

	"firewall/internal/rules"
)

func ApplyRule(r rules.Rule) error {
	_ = r
	return fmt.Errorf("linux adapter not available on this platform")
}
