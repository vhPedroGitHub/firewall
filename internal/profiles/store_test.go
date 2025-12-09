package profiles

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestStore(t *testing.T) *SQLiteStore {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	store, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return store
}

func TestProfileStore_SaveAndGet(t *testing.T) {
	store := setupTestStore(t)

	p := Profile{
		Name:        "work",
		Description: "Work profile",
		Active:      false,
		Rules:       []string{"web"},
	}

	if err := store.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	got, err := store.GetProfile("work")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if got.Name != p.Name {
		t.Errorf("expected name %q, got %q", p.Name, got.Name)
	}
	if got.Description != p.Description {
		t.Errorf("expected description %q, got %q", p.Description, got.Description)
	}
	if got.Active != p.Active {
		t.Errorf("expected active %v, got %v", p.Active, got.Active)
	}
	if len(got.Rules) != len(p.Rules) {
		t.Fatalf("expected %d rules, got %d", len(p.Rules), len(got.Rules))
	}
	if got.Rules[0] != p.Rules[0] {
		t.Errorf("expected rule name %q, got %q", p.Rules[0], got.Rules[0])
	}
}

func TestProfileStore_List(t *testing.T) {
	store := setupTestStore(t)

	profiles := []Profile{
		{Name: "work", Description: "Work profile", Active: false, Rules: []string{}},
		{Name: "home", Description: "Home profile", Active: false, Rules: []string{}},
	}

	for _, p := range profiles {
		if err := store.SaveProfile(p); err != nil {
			t.Fatalf("SaveProfile failed: %v", err)
		}
	}

	list, err := store.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(list))
	}

	// Check sorted order (should be alphabetical by name)
	if list[0].Name != "home" || list[1].Name != "work" {
		t.Errorf("profiles not in expected order: got %v", list)
	}
}

func TestProfileStore_Delete(t *testing.T) {
	store := setupTestStore(t)

	p := Profile{Name: "temp", Description: "Temporary", Active: false, Rules: []string{}}
	if err := store.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	if err := store.DeleteProfile("temp"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	_, err := store.GetProfile("temp")
	if err == nil {
		t.Fatal("expected error getting deleted profile, got nil")
	}
}

func TestProfileStore_SetActive(t *testing.T) {
	store := setupTestStore(t)

	profiles := []Profile{
		{Name: "work", Description: "Work", Active: false, Rules: []string{}},
		{Name: "home", Description: "Home", Active: false, Rules: []string{}},
	}

	for _, p := range profiles {
		if err := store.SaveProfile(p); err != nil {
			t.Fatalf("SaveProfile failed: %v", err)
		}
	}

	if err := store.SetActiveProfile("work"); err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}

	active, err := store.GetActiveProfile()
	if err != nil {
		t.Fatalf("GetActiveProfile failed: %v", err)
	}

	if active.Name != "work" {
		t.Errorf("expected active profile %q, got %q", "work", active.Name)
	}
	if !active.Active {
		t.Error("expected Active to be true")
	}

	// Set different profile active
	if err := store.SetActiveProfile("home"); err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}

	active, err = store.GetActiveProfile()
	if err != nil {
		t.Fatalf("GetActiveProfile failed: %v", err)
	}

	if active.Name != "home" {
		t.Errorf("expected active profile %q, got %q", "home", active.Name)
	}

	// Verify work is no longer active
	work, err := store.GetProfile("work")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}
	if work.Active {
		t.Error("expected work profile to be inactive")
	}
}

func TestProfileStore_Validation(t *testing.T) {
	store := setupTestStore(t)

	tests := []struct {
		name    string
		profile Profile
		wantErr bool
	}{
		{
			name:    "missing name",
			profile: Profile{Description: "Test", Rules: []string{}},
			wantErr: true,
		},
		{
			name:    "missing description",
			profile: Profile{Name: "test", Rules: []string{}},
			wantErr: true,
		},
		{
			name:    "valid profile",
			profile: Profile{Name: "test", Description: "Test profile", Rules: []string{}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.SaveProfile(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileStore_Update(t *testing.T) {
	store := setupTestStore(t)

	p := Profile{
		Name:        "test",
		Description: "Original description",
		Active:      false,
		Rules:       []string{},
	}

	if err := store.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	// Update description and rules
	p.Description = "Updated description"
	p.Rules = []string{"ssh"}

	if err := store.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile (update) failed: %v", err)
	}

	got, err := store.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if got.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", got.Description)
	}
	if len(got.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got.Rules))
	}
	if got.Rules[0] != "ssh" {
		t.Errorf("expected rule name %q, got %q", "ssh", got.Rules[0])
	}
}
