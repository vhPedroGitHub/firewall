package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestRulesCommands_AddListRemove(t *testing.T) {
	dir := t.TempDir()
	dbPath = filepath.Join(dir, "rules.db")

	// reset global state in case other tests run
	db = nil
	ruleStore = nil

	// Add rule
	out, err := runCLI("rules", "add",
		"--name", "web",
		"--app", "C:/App/app.exe",
		"--action", "allow",
		"--protocol", "tcp",
		"--direction", "outbound",
		"--ports", "80,443",
	)
	if err != nil {
		t.Fatalf("add rule: %v", err)
	}
	if !contains(out, "rule \"web\" saved") {
		t.Fatalf("unexpected output: %s", out)
	}

	// List rules
	out, err = runCLI("rules", "list")
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if !contains(out, "web [allow tcp outbound]") {
		t.Fatalf("list output missing rule: %s", out)
	}

	// Remove rule
	out, err = runCLI("rules", "remove", "--name", "web")
	if err != nil {
		t.Fatalf("remove rule: %v", err)
	}
	if !contains(out, "rule \"web\" removed") {
		t.Fatalf("unexpected remove output: %s", out)
	}
}

func runCLI(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// ensure cleanup after each command
	defer func() {
		cleanupStore(nil, nil)
	}()

	err := rootCmd.Execute()
	return buf.String(), err
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
