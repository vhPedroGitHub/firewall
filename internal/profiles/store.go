package profiles

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Store defines persistence operations for profiles.
type Store interface {
	ListProfiles() ([]Profile, error)
	SaveProfile(profile Profile) error
	DeleteProfile(name string) error
	GetProfile(name string) (*Profile, error)
	SetActiveProfile(name string) error
	GetActiveProfile() (*Profile, error)
}

// SQLiteStore is a sqlite-backed implementation of Store.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a sqlite-backed profile store.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if err := initSchema(db); err != nil {
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS profiles (
	name TEXT PRIMARY KEY,
	description TEXT NOT NULL,
	active INTEGER NOT NULL DEFAULT 0,
	rules TEXT NOT NULL
);
`
	_, err := db.Exec(schema)
	return err
}

// ListProfiles lists all profiles.
func (s *SQLiteStore) ListProfiles() ([]Profile, error) {
	rows, err := s.db.Query(`SELECT name, description, active, rules FROM profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Profile
	for rows.Next() {
		var p Profile
		var active int
		var rulesJSON string
		if err := rows.Scan(&p.Name, &p.Description, &active, &rulesJSON); err != nil {
			return nil, err
		}
		p.Active = active == 1
		if err := json.Unmarshal([]byte(rulesJSON), &p.Rules); err != nil {
			return nil, fmt.Errorf("unmarshal rules for %s: %w", p.Name, err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveProfile validates and persists a profile.
func (s *SQLiteStore) SaveProfile(profile Profile) error {
	if err := Validate(profile); err != nil {
		return err
	}
	rulesJSON, err := json.Marshal(profile.Rules)
	if err != nil {
		return err
	}
	active := 0
	if profile.Active {
		active = 1
	}
	_, err = s.db.Exec(`INSERT OR REPLACE INTO profiles (name, description, active, rules) VALUES (?,?,?,?)`,
		profile.Name, profile.Description, active, string(rulesJSON))
	return err
}

// DeleteProfile removes a profile by name.
func (s *SQLiteStore) DeleteProfile(name string) error {
	_, err := s.db.Exec(`DELETE FROM profiles WHERE name = ?`, name)
	return err
}

// GetProfile retrieves a profile by name.
func (s *SQLiteStore) GetProfile(name string) (*Profile, error) {
	var p Profile
	var active int
	var rulesJSON string
	err := s.db.QueryRow(`SELECT name, description, active, rules FROM profiles WHERE name = ?`, name).
		Scan(&p.Name, &p.Description, &active, &rulesJSON)
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	if err := json.Unmarshal([]byte(rulesJSON), &p.Rules); err != nil {
		return nil, err
	}
	return &p, nil
}

// SetActiveProfile sets a profile as active (deactivates others).
func (s *SQLiteStore) SetActiveProfile(name string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deactivate all
	if _, err := tx.Exec(`UPDATE profiles SET active = 0`); err != nil {
		return err
	}
	// Activate target
	if _, err := tx.Exec(`UPDATE profiles SET active = 1 WHERE name = ?`, name); err != nil {
		return err
	}
	return tx.Commit()
}

// GetActiveProfile retrieves the currently active profile.
func (s *SQLiteStore) GetActiveProfile() (*Profile, error) {
	var p Profile
	var active int
	var rulesJSON string
	err := s.db.QueryRow(`SELECT name, description, active, rules FROM profiles WHERE active = 1 LIMIT 1`).
		Scan(&p.Name, &p.Description, &active, &rulesJSON)
	if err != nil {
		return nil, err
	}
	p.Active = true
	if err := json.Unmarshal([]byte(rulesJSON), &p.Rules); err != nil {
		return nil, err
	}
	return &p, nil
}
