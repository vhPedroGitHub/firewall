package monitor

import (
	"fmt"
	"strings"

	"github.com/vhPedroGitHub/firewall/internal/notify"
	"github.com/vhPedroGitHub/firewall/internal/rules"
)

// DefaultHandler implements Handler with rule checking and user prompts.
type DefaultHandler struct {
	Store rules.Store
}

// NewDefaultHandler creates a new handler with the given rule store.
func NewDefaultHandler(store rules.Store) *DefaultHandler {
	return &DefaultHandler{Store: store}
}

// HandleConnection checks if a rule exists for the connection event.
// If no rule exists, it prompts the user and returns their decision.
func (h *DefaultHandler) HandleConnection(event ConnectionEvent) (Decision, error) {
	return h.HandleConnectionWithPrompts(event, true)
}

// HandleConnectionWithPrompts checks if a rule exists for the connection event.
// If no rule exists and prompts are enabled, it prompts the user.
// If prompts are disabled, it denies the connection.
func (h *DefaultHandler) HandleConnectionWithPrompts(event ConnectionEvent, promptsEnabled bool) (Decision, error) {
	// Check if we have a matching rule
	existingRules, err := h.Store.ListRules()
	if err != nil {
		return DecisionDeny, fmt.Errorf("failed to list rules: %w", err)
	}

	for _, rule := range existingRules {
		if h.matchesRule(event, rule) {
			// Rule found - apply it
			if rule.Action == "allow" {
				return DecisionAllow, nil
			}
			return DecisionDeny, nil
		}
	}

	// No matching rule found
	if !promptsEnabled {
		// Prompts disabled - deny by default
		return DecisionDeny, nil
	}

	// Prompts enabled - ask user
	decision, err := h.promptUser(event)
	if err != nil {
		return DecisionDeny, fmt.Errorf("failed to prompt user: %w", err)
	}

	return decision, nil
}

// matchesRule checks if a connection event matches a rule.
func (h *DefaultHandler) matchesRule(event ConnectionEvent, rule rules.Rule) bool {
	// Check app path
	if rule.Application != "" && !strings.EqualFold(rule.Application, event.AppPath) {
		return false
	}

	// Check protocol
	if rule.Protocol != "" && !strings.EqualFold(rule.Protocol, event.Protocol) {
		return false
	}

	// Check direction
	if rule.Direction != "" && !strings.EqualFold(rule.Direction, event.Direction) {
		return false
	}

	// Check ports if specified
	if len(rule.Ports) > 0 {
		portMatch := false
		eventPort := event.DstPort
		if event.Direction == "inbound" {
			eventPort = event.SrcPort
		}

		for _, p := range rule.Ports {
			if p == eventPort {
				portMatch = true
				break
			}
		}
		if !portMatch {
			return false
		}
	}

	return true
}

// promptUser displays a notification asking the user to allow or deny the connection.
func (h *DefaultHandler) promptUser(event ConnectionEvent) (Decision, error) {
	// Build prompt message
	msg := fmt.Sprintf(
		"Application: %s\nProtocol: %s\nDirection: %s\nFrom: %s:%d\nTo: %s:%d\n\nAllow this connection?",
		event.AppPath,
		event.Protocol,
		event.Direction,
		event.SrcAddr,
		event.SrcPort,
		event.DstAddr,
		event.DstPort,
	)

	// Show notification with Yes/No options
	result, err := notify.Show("Firewall Connection Request", msg)
	if err != nil {
		return DecisionDeny, err
	}

	// Parse result - notify.Show returns "yes" or "no" from user dialog
	result = strings.ToLower(strings.TrimSpace(result))
	switch result {
	case "yes", "allow", "ok":
		return DecisionAllow, nil
	case "no", "deny", "cancel":
		return DecisionDeny, nil
	default:
		return DecisionCancel, nil
	}
}

// SaveDecisionAsRule saves a user's decision as a permanent rule.
func (h *DefaultHandler) SaveDecisionAsRule(event ConnectionEvent, decision Decision) error {
	if decision == DecisionCancel {
		return nil // Don't save cancelled decisions
	}

	action := "deny"
	if decision == DecisionAllow {
		action = "allow"
	}

	// Generate a unique rule name
	ruleName := fmt.Sprintf("auto_%s_%s_%d",
		sanitizeForRuleName(event.AppPath),
		event.Protocol,
		event.DstPort,
	)

	rule := rules.Rule{
		Name:        ruleName,
		Application: event.AppPath,
		Action:      action,
		Protocol:    event.Protocol,
		Direction:   event.Direction,
		Ports:       []int{event.DstPort},
	}

	if err := rules.Validate(rule); err != nil {
		return fmt.Errorf("generated invalid rule: %w", err)
	}

	return h.Store.SaveRule(rule)
}

// sanitizeForRuleName converts a file path to a safe rule name component.
func sanitizeForRuleName(path string) string {
	// Extract just the filename - handle both Windows and Unix paths
	filename := path

	// Try Windows path separator first
	if idx := strings.LastIndex(path, "\\"); idx >= 0 {
		filename = path[idx+1:]
	} else if idx := strings.LastIndex(path, "/"); idx >= 0 {
		// Try Unix path separator
		filename = path[idx+1:]
	}

	// Remove common extensions
	filename = strings.TrimSuffix(filename, ".exe")

	// Replace invalid characters
	replacer := strings.NewReplacer(
		" ", "_",
		".", "_",
		"-", "_",
		"(", "",
		")", "",
	)
	return replacer.Replace(filename)
}

// CheckRule implements RuleChecker interface.
func (h *DefaultHandler) CheckRule(event ConnectionEvent) *rules.Rule {
	existingRules, err := h.Store.ListRules()
	if err != nil {
		return nil
	}

	for _, rule := range existingRules {
		if h.matchesRule(event, rule) {
			return &rule
		}
	}
	return nil
}
