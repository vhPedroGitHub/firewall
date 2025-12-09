package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"firewall/internal/profiles"
)

var (
	profileName        string
	profileDescription string
	profileExportPath  string
	profileImportPath  string
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage firewall profiles",
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileStore == nil {
			return errors.New("profile store not initialized")
		}
		list, err := profileStore.ListProfiles()
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no profiles configured")
			return nil
		}
		for _, p := range list {
			active := ""
			if p.Active {
				active = " (active)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "- %s%s: %s [%d rules]\n", p.Name, active, p.Description, len(p.Rules))
		}
		return nil
	},
}

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileStore == nil {
			return errors.New("profile store not initialized")
		}
		p := profiles.Profile{
			Name:        profileName,
			Description: profileDescription,
			Rules:       []string{},
		}
		if err := profileStore.SaveProfile(p); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "profile %q created\n", p.Name)
		return nil
	},
}

var profilesActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileStore == nil {
			return errors.New("profile store not initialized")
		}
		if profileName == "" {
			return errors.New("--name is required")
		}
		if err := profileStore.SetActiveProfile(profileName); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "profile %q activated\n", profileName)
		return nil
	},
}

var profilesExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a profile to JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileStore == nil {
			return errors.New("profile store not initialized")
		}
		if profileName == "" {
			return errors.New("--name is required")
		}
		p, err := profileStore.GetProfile(profileName)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		if profileExportPath == "" {
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
		} else {
			if err := os.WriteFile(profileExportPath, data, 0644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "profile %q exported to %s\n", profileName, profileExportPath)
		}
		return nil
	},
}

var profilesImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a profile from JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if profileStore == nil {
			return errors.New("profile store not initialized")
		}
		if profileImportPath == "" {
			return errors.New("--file is required")
		}
		data, err := os.ReadFile(profileImportPath)
		if err != nil {
			return err
		}
		var p profiles.Profile
		if err := json.Unmarshal(data, &p); err != nil {
			return err
		}
		if err := profileStore.SaveProfile(p); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "profile %q imported\n", p.Name)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesActivateCmd)
	profilesCmd.AddCommand(profilesExportCmd)
	profilesCmd.AddCommand(profilesImportCmd)

	profilesCreateCmd.Flags().StringVar(&profileName, "name", "", "profile name (required)")
	profilesCreateCmd.Flags().StringVar(&profileDescription, "description", "", "profile description")
	_ = profilesCreateCmd.MarkFlagRequired("name")

	profilesActivateCmd.Flags().StringVar(&profileName, "name", "", "profile name (required)")
	_ = profilesActivateCmd.MarkFlagRequired("name")

	profilesExportCmd.Flags().StringVar(&profileName, "name", "", "profile name (required)")
	profilesExportCmd.Flags().StringVar(&profileExportPath, "file", "", "export file path (optional, prints to stdout if omitted)")
	_ = profilesExportCmd.MarkFlagRequired("name")

	profilesImportCmd.Flags().StringVar(&profileImportPath, "file", "", "import file path (required)")
	_ = profilesImportCmd.MarkFlagRequired("file")
}
