package app

import (
	"fmt"

	"firewall/internal/logging"
	"firewall/internal/platform"
	"firewall/internal/rules"
)

// Service centralizes core operations shared by CLI and GUI.
type Service struct {
	Store rules.Store
}

// ListRules returns stored rules.
func (s *Service) ListRules() ([]rules.Rule, error) {
	return s.Store.ListRules()
}

// SaveRule validates and persists a rule.
func (s *Service) SaveRule(r rules.Rule) error {
	if err := s.Store.SaveRule(r); err != nil {
		return err
	}
	logging.LogEvent("info", "rule-add", fmt.Sprintf("Rule %q saved via service", r.Name), map[string]interface{}{
		"name":      r.Name,
		"app":       r.Application,
		"action":    r.Action,
		"protocol":  r.Protocol,
		"direction": r.Direction,
		"ports":     r.Ports,
	})
	return nil
}

// DeleteRule removes a rule by name.
func (s *Service) DeleteRule(name string) error {
	if err := s.Store.DeleteRule(name); err != nil {
		return err
	}
	logging.LogEvent("info", "rule-remove", fmt.Sprintf("Rule %q deleted via service", name), map[string]interface{}{
		"name": name,
	})
	return nil
}

// ApplyRule dispatches to the platform-specific adapter.
func (s *Service) ApplyRule(r rules.Rule) error {
	return platform.ApplyRule(r)
}
