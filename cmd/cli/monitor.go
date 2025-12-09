package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vhPedroGitHub/firewall/internal/monitor"
)

var (
	monitorSvc *monitor.Service
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Control connection monitoring",
	Long:  `Start or stop monitoring network connections for automatic rule prompts.`,
}

var monitorStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start connection monitoring",
	Long:  `Start monitoring network connections. When an unknown connection is detected, you will be prompted to allow or deny it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorSvc == nil {
			var err error
			monitorSvc, err = monitor.NewService(ruleStore)
			if err != nil {
				return fmt.Errorf("failed to create monitor service: %w", err)
			}
		}

		if err := monitorSvc.Start(); err != nil {
			return fmt.Errorf("failed to start monitoring: %w", err)
		}

		fmt.Println("Connection monitoring started. Press Ctrl+C to stop.")

		// Keep the program running
		select {}
	},
}

var monitorStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop connection monitoring",
	Long:  `Stop monitoring network connections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorSvc == nil {
			return fmt.Errorf("monitor service not running")
		}

		if err := monitorSvc.Stop(); err != nil {
			return fmt.Errorf("failed to stop monitoring: %w", err)
		}

		fmt.Println("Connection monitoring stopped.")
		return nil
	},
}

var monitorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show monitoring status",
	Long:  `Display whether connection monitoring is currently active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorSvc == nil || !monitorSvc.IsRunning() {
			fmt.Println("Status: Stopped")
		} else {
			fmt.Println("Status: Running")
		}
		return nil
	},
}

func init() {
	monitorCmd.AddCommand(monitorStartCmd)
	monitorCmd.AddCommand(monitorStopCmd)
	monitorCmd.AddCommand(monitorStatusCmd)
	rootCmd.AddCommand(monitorCmd)
}
