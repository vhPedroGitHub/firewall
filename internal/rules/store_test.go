package rules

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteStore_CRUD(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	store, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	rule := Rule{
		Name:        "web",
		Application: "/usr/bin/app",
		Action:      "allow",
		Protocol:    "tcp",
		Direction:   "outbound",
		Ports:       []int{80, 443},
	}

	if err := store.SaveRule(rule); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.ListRules()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}
	if got[0].Name != rule.Name || len(got[0].Ports) != len(rule.Ports) {
		t.Fatalf("unexpected rule: %+v", got[0])
	}

	if err := store.DeleteRule(rule.Name); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, err = store.ListRules()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 rules after delete, got %d", len(got))
	}
}

func TestPortHelpers(t *testing.T) {
	ports := []int{80, 443}
	joined := joinPorts(ports)
	parsed, err := parsePorts(joined)
	if err != nil {
		t.Fatalf("parsePorts: %v", err)
	}
	if len(parsed) != len(ports) || parsed[0] != 80 || parsed[1] != 443 {
		t.Fatalf("unexpected parsed ports: %v", parsed)
	}
}
