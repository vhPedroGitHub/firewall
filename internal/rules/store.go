package rules

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Store defines minimal persistence operations for firewall rules.
type Store interface {
	ListRules() ([]Rule, error)
	SaveRule(rule Rule) error
	DeleteRule(name string) error
}

// SQLiteStore is a sqlite-backed implementation of Store.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore wires a sqlite-backed rule store; caller owns DB lifecycle.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if err := initSchema(db); err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS rules (
	name TEXT PRIMARY KEY,
	application TEXT NOT NULL,
	action TEXT NOT NULL,
	protocol TEXT NOT NULL,
	direction TEXT NOT NULL,
	ports TEXT NOT NULL
);
`
	_, err := db.Exec(schema)
	return err
}

// ListRules lists rules from sqlite.
func (s *SQLiteStore) ListRules() ([]Rule, error) {
	rows, err := s.db.Query(`SELECT name, application, action, protocol, direction, ports FROM rules ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Rule
	for rows.Next() {
		var r Rule
		var ports string
		if err := rows.Scan(&r.Name, &r.Application, &r.Action, &r.Protocol, &r.Direction, &ports); err != nil {
			return nil, err
		}
		parsed, err := parsePorts(ports)
		if err != nil {
			return nil, fmt.Errorf("invalid stored ports for %s: %w", r.Name, err)
		}
		r.Ports = parsed
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveRule validates and persists a rule (upsert).
func (s *SQLiteStore) SaveRule(rule Rule) error {
	if err := Validate(rule); err != nil {
		return err
	}
	ports := joinPorts(rule.Ports)
	_, err := s.db.Exec(`INSERT OR REPLACE INTO rules (name, application, action, protocol, direction, ports) VALUES (?,?,?,?,?,?)`,
		rule.Name, rule.Application, rule.Action, rule.Protocol, rule.Direction, ports)
	return err
}

// DeleteRule removes a rule by name.
func (s *SQLiteStore) DeleteRule(name string) error {
	_, err := s.db.Exec(`DELETE FROM rules WHERE name = ?`, name)
	return err
}

func joinPorts(ports []int) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		parts = append(parts, strconv.Itoa(p))
	}
	return strings.Join(parts, ",")
}

func parsePorts(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	chunks := strings.Split(raw, ",")
	out := make([]int, 0, len(chunks))
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		v, err := strconv.Atoi(c)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
