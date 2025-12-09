package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.1-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("firewall CLI %s\n", version)
		return nil
	},
}
