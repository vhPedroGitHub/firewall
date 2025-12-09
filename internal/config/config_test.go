package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Default(t *testing.T) {
	cfg := Default()

	if cfg.DBPath != "firewall.db" {
		t.Errorf("expected DBPath %q, got %q", "firewall.db", cfg.DBPath)
	}
	if cfg.LogPath != "firewall.log" {
		t.Errorf("expected LogPath %q, got %q", "firewall.log", cfg.LogPath)
	}
	if cfg.GUI.Width != 1024 {
		t.Errorf("expected GUI Width 1024, got %d", cfg.GUI.Width)
	}
	if cfg.GUI.Height != 768 {
		t.Errorf("expected GUI Height 768, got %d", cfg.GUI.Height)
	}
}

func TestConfig_LoadNonexistent(t *testing.T) {
	cfg, err := Load("nonexistent-config.json")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should return defaults
	def := Default()
	if cfg.DBPath != def.DBPath {
		t.Errorf("expected default DBPath, got %q", cfg.DBPath)
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Config{
		DBPath:         "/custom/firewall.db",
		LogPath:        "/custom/firewall.log",
		DefaultProfile: "work",
		GUI: GUIConfig{
			Width:  1920,
			Height: 1080,
			Theme:  "light",
		},
	}

	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.DBPath != cfg.DBPath {
		t.Errorf("expected DBPath %q, got %q", cfg.DBPath, loaded.DBPath)
	}
	if loaded.LogPath != cfg.LogPath {
		t.Errorf("expected LogPath %q, got %q", cfg.LogPath, loaded.LogPath)
	}
	if loaded.DefaultProfile != cfg.DefaultProfile {
		t.Errorf("expected DefaultProfile %q, got %q", cfg.DefaultProfile, loaded.DefaultProfile)
	}
	if loaded.GUI.Width != cfg.GUI.Width {
		t.Errorf("expected GUI Width %d, got %d", cfg.GUI.Width, loaded.GUI.Width)
	}
	if loaded.GUI.Theme != cfg.GUI.Theme {
		t.Errorf("expected GUI Theme %q, got %q", cfg.GUI.Theme, loaded.GUI.Theme)
	}
}

func TestConfig_LoadWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "partial.json")

	// Write partial config
	partial := `{"db_path": "/tmp/test.db"}`
	if err := os.WriteFile(cfgPath, []byte(partial), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.DBPath != "/tmp/test.db" {
		t.Errorf("expected custom DBPath, got %q", loaded.DBPath)
	}

	// Other fields should have defaults
	def := Default()
	if loaded.LogPath != def.LogPath {
		t.Errorf("expected default LogPath, got %q", loaded.LogPath)
	}
	if loaded.GUI.Width != def.GUI.Width {
		t.Errorf("expected default GUI Width, got %d", loaded.GUI.Width)
	}
}

func TestConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(cfgPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
