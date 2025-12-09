package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"

	"github.com/vhPedroGitHub/firewall/internal/config"
	"github.com/vhPedroGitHub/firewall/internal/logging"
	"github.com/vhPedroGitHub/firewall/internal/profiles"
	"github.com/vhPedroGitHub/firewall/internal/rules"
)

var (
	cfgPath      string
	verbose      bool
	dbPath       string
	db           *sql.DB
	ruleStore    rules.Store
	profileStore profiles.Store
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Cross-platform firewall utility CLI",
	Long:  "Manage firewall rules, profiles, and diagnostics via CLI.",
}

// Execute runs the Cobra root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config file (optional)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "firewall.db", "path to sqlite database")

	rootCmd.PersistentPreRunE = ensureStore
	rootCmd.PersistentPostRun = cleanupStore

	rootCmd.AddCommand(rulesCmd)
	rootCmd.AddCommand(profilesCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// Load config file if specified, otherwise use defaults
	if cfgPath == "" {
		cfgPath = "firewall.json"
	}
	_, _ = config.Load(cfgPath) // Ignore errors, use defaults
}

// Not yet wired; placeholder for future config/env loading.
func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func ensureStore(cmd *cobra.Command, args []string) error {
	if ruleStore != nil {
		return nil
	}

	// Load config
	cfg, _ := config.Load(cfgPath)
	if dbPath == "firewall.db" { // Use config if flag not overridden
		dbPath = cfg.DBPath
	}

	handle, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	store, err := rules.NewSQLiteStore(handle)
	if err != nil {
		handle.Close()
		return err
	}
	pStore, err := profiles.NewSQLiteStore(handle)
	if err != nil {
		handle.Close()
		return err
	}
	// Initialize logging
	logPath := cfg.LogPath
	if logPath == "" {
		logPath = "firewall.log"
	}
	if err := logging.Init(logPath); err != nil {
		handle.Close()
		return err
	}
	db = handle
	ruleStore = store
	profileStore = pStore
	return nil
}

func cleanupStore(cmd *cobra.Command, args []string) {
	if db != nil {
		_ = db.Close()
		db = nil
	}
	ruleStore = nil
	profileStore = nil
	_ = logging.Close()
}
