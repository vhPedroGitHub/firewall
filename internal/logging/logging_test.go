package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogger_LogEvent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	LogEvent("info", "test-category", "test message", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	})

	// Verify file was created and has content
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Fatal("expected log content, got empty file")
	}

	// Check for expected fields
	if !contains(content, "test-category") {
		t.Error("log missing category")
	}
	if !contains(content, "test message") {
		t.Error("log missing message")
	}
	if !contains(content, "value1") {
		t.Error("log missing detail value")
	}
}

func TestLogger_MultipleEvents(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	LogEvent("info", "event1", "first event", nil)
	LogEvent("warning", "event2", "second event", nil)
	LogEvent("error", "event3", "third event", nil)

	events, err := ReadEvents(logPath)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Level != "info" || events[0].Category != "event1" {
		t.Errorf("first event mismatch: %+v", events[0])
	}
	if events[1].Level != "warning" || events[1].Category != "event2" {
		t.Errorf("second event mismatch: %+v", events[1])
	}
	if events[2].Level != "error" || events[2].Category != "event3" {
		t.Errorf("third event mismatch: %+v", events[2])
	}
}

func TestLogger_ReadEvents(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "read.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	details := map[string]interface{}{
		"name":   "test-rule",
		"action": "allow",
		"port":   443,
	}

	LogEvent("info", "rule-add", "Rule added", details)
	Close()

	events, err := ReadEvents(logPath)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	e := events[0]
	if e.Level != "info" {
		t.Errorf("expected level 'info', got %q", e.Level)
	}
	if e.Category != "rule-add" {
		t.Errorf("expected category 'rule-add', got %q", e.Category)
	}
	if e.Message != "Rule added" {
		t.Errorf("expected message 'Rule added', got %q", e.Message)
	}
	if e.Details["name"] != "test-rule" {
		t.Errorf("expected detail name 'test-rule', got %v", e.Details["name"])
	}
	if e.Details["port"] != float64(443) { // JSON numbers unmarshal as float64
		t.Errorf("expected detail port 443, got %v", e.Details["port"])
	}
}

func TestLogger_Timestamp(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "timestamp.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	before := time.Now()
	LogEvent("info", "test", "timestamp test", nil)
	after := time.Now()

	events, err := ReadEvents(logPath)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	ts := events[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestLogger_NoInit(t *testing.T) {
	// LogEvent should not panic when logger not initialized
	LogEvent("info", "test", "should be ignored", nil)
}

func TestLogger_EmptyDetails(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "empty.log")

	if err := Init(logPath); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	LogEvent("info", "test", "no details", nil)

	events, err := ReadEvents(logPath)
	if err != nil {
		t.Fatalf("ReadEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Details != nil {
		t.Errorf("expected nil details, got %v", events[0].Details)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexStr(s, substr) >= 0)
}

func indexStr(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
