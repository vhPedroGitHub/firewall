package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"firewall/internal/logging"
	"firewall/internal/rules"
)

var (
	addName      string
	addApp       string
	addAction    string
	addProtocol  string
	addDirection string
	addPorts     string
	removeName   string
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage firewall rules",
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ruleStore == nil {
			return errors.New("rule store not initialized")
		}
		list, err := ruleStore.ListRules()
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no rules configured")
			return nil
		}
		for _, r := range list {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s [%s %s %s] app=%s ports=%v\n", r.Name, r.Action, r.Protocol, r.Direction, r.Application, r.Ports)
		}
		return nil
	},
}

var rulesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a firewall rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ruleStore == nil {
			return errors.New("rule store not initialized")
		}
		ports, err := parsePortsFlag(addPorts)
		if err != nil {
			return err
		}
		r := rules.Rule{
			Name:        addName,
			Application: addApp,
			Action:      addAction,
			Protocol:    addProtocol,
			Direction:   addDirection,
			Ports:       ports,
		}
		if err := ruleStore.SaveRule(r); err != nil {
			return err
		}
		logging.LogEvent("info", "rule-add", fmt.Sprintf("Rule %q added", r.Name), map[string]interface{}{
			"name":      r.Name,
			"app":       r.Application,
			"action":    r.Action,
			"protocol":  r.Protocol,
			"direction": r.Direction,
			"ports":     r.Ports,
		})
		fmt.Fprintf(cmd.OutOrStdout(), "rule %q saved\n", r.Name)
		return nil
	},
}

var rulesRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a firewall rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ruleStore == nil {
			return errors.New("rule store not initialized")
		}
		if removeName == "" {
			return errors.New("--name is required")
		}
		if err := ruleStore.DeleteRule(removeName); err != nil {
			return err
		}
		logging.LogEvent("info", "rule-remove", fmt.Sprintf("Rule %q removed", removeName), map[string]interface{}{
			"name": removeName,
		})
		fmt.Fprintf(cmd.OutOrStdout(), "rule %q removed\n", removeName)
		return nil
	},
}

func parsePortsFlag(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", p)
		}
		out = append(out, v)
	}
	return out, nil
}

func init() {
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesAddCmd)
	rulesCmd.AddCommand(rulesRemoveCmd)

	rulesAddCmd.Flags().StringVar(&addName, "name", "", "rule name (required)")
	rulesAddCmd.Flags().StringVar(&addApp, "app", "", "application path or identifier (required)")
	rulesAddCmd.Flags().StringVar(&addAction, "action", "allow", "action: allow or deny")
	rulesAddCmd.Flags().StringVar(&addProtocol, "protocol", "tcp", "protocol: tcp|udp|any")
	rulesAddCmd.Flags().StringVar(&addDirection, "direction", "outbound", "direction: inbound|outbound")
	rulesAddCmd.Flags().StringVar(&addPorts, "ports", "", "comma-separated port list (required for tcp/udp)")

	_ = rulesAddCmd.MarkFlagRequired("name")
	_ = rulesAddCmd.MarkFlagRequired("app")

	rulesRemoveCmd.Flags().StringVar(&removeName, "name", "", "rule name to remove (required)")
	_ = rulesRemoveCmd.MarkFlagRequired("name")
}
