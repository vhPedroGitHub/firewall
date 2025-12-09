package monitor

import (
	"testing"
	"time"

	"github.com/vhPedroGitHub/firewall/internal/rules"
)

// mockStore is a mock implementation of rules.Store for testing
type mockStore struct {
	rules []rules.Rule
}

func (m *mockStore) SaveRule(rule rules.Rule) error {
	m.rules = append(m.rules, rule)
	return nil
}

func (m *mockStore) ListRules() ([]rules.Rule, error) {
	return m.rules, nil
}

func (m *mockStore) DeleteRule(name string) error {
	for i, r := range m.rules {
		if r.Name == name {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockStore) GetRule(name string) (*rules.Rule, error) {
	for _, r := range m.rules {
		if r.Name == name {
			return &r, nil
		}
	}
	return nil, nil
}

func TestService_EventTracking(t *testing.T) {
	store := &mockStore{}
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Create test event
	event := ConnectionEvent{
		AppPath:   "/usr/bin/test",
		PID:       "1234",
		Protocol:  "tcp",
		Direction: "outbound",
		SrcAddr:   "192.168.1.100",
		SrcPort:   50000,
		DstAddr:   "93.184.216.34",
		DstPort:   443,
		State:     "ESTABLISHED",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	eventLog := ConnectionEventLog{
		Event:     event,
		Decision:  "allowed",
		Timestamp: time.Now(),
		RuleName:  "test_rule",
	}

	// Add event log
	svc.addEventLog(eventLog)

	// Verify event was added
	events := svc.GetRecentEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Decision != "allowed" {
		t.Errorf("Expected decision 'allowed', got '%s'", events[0].Decision)
	}

	// Test clear events
	svc.ClearEvents()
	events = svc.GetRecentEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events after clear, got %d", len(events))
	}
}

func TestService_PromptsControl(t *testing.T) {
	store := &mockStore{}
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Prompts should be enabled by default
	if !svc.PromptsEnabled() {
		t.Error("Prompts should be enabled by default")
	}

	// Disable prompts
	svc.DisablePrompts()
	if svc.PromptsEnabled() {
		t.Error("Prompts should be disabled")
	}

	// Re-enable prompts
	svc.EnablePrompts()
	if !svc.PromptsEnabled() {
		t.Error("Prompts should be enabled")
	}
}

func TestService_TrafficTracking(t *testing.T) {
	store := &mockStore{}
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	event := ConnectionEvent{
		AppPath:   "/usr/bin/test",
		Protocol:  "tcp",
		Direction: "outbound",
		SrcAddr:   "192.168.1.100",
		SrcPort:   50000,
		DstAddr:   "93.184.216.34",
		DstPort:   443,
	}

	// Track some traffic
	svc.trackTraffic(event, 1024)
	svc.trackTraffic(event, 2048)

	// Get traffic stats
	traffic := svc.GetProcessTraffic()
	if len(traffic) != 1 {
		t.Fatalf("Expected 1 traffic entry, got %d", len(traffic))
	}

	if traffic[0].AppPath != "/usr/bin/test" {
		t.Errorf("Expected app path '/usr/bin/test', got '%s'", traffic[0].AppPath)
	}

	if traffic[0].BytesSent != 3072 {
		t.Errorf("Expected 3072 bytes sent, got %d", traffic[0].BytesSent)
	}

	if traffic[0].Connections != 2 {
		t.Errorf("Expected 2 connections, got %d", traffic[0].Connections)
	}

	// Test clear traffic
	svc.ClearProcessTraffic()
	traffic = svc.GetProcessTraffic()
	if len(traffic) != 0 {
		t.Errorf("Expected 0 traffic entries after clear, got %d", len(traffic))
	}
}

func TestService_ActiveProcesses(t *testing.T) {
	store := &mockStore{}
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Initially no processes
	processes := svc.GetActiveProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected 0 active processes, got %d", len(processes))
	}

	// Manually add a process (simulating what processEvents would do)
	event := ConnectionEvent{
		AppPath: "/usr/bin/test",
		PID:     "1234",
	}
	svc.processesMu.Lock()
	svc.activeProcesses[event.AppPath] = event
	svc.processesMu.Unlock()

	// Verify process was added
	processes = svc.GetActiveProcesses()
	if len(processes) != 1 {
		t.Errorf("Expected 1 active process, got %d", len(processes))
	}

	// Clear processes
	svc.ClearActiveProcesses()
	processes = svc.GetActiveProcesses()
	if len(processes) != 0 {
		t.Errorf("Expected 0 active processes after clear, got %d", len(processes))
	}
}

func TestService_EventLogLimit(t *testing.T) {
	store := &mockStore{}
	svc, err := NewService(store)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Add more events than maxEvents
	for i := 0; i < 150; i++ {
		eventLog := ConnectionEventLog{
			Event: ConnectionEvent{
				AppPath: "/usr/bin/test",
				DstPort: i,
			},
			Decision:  "allowed",
			Timestamp: time.Now(),
		}
		svc.addEventLog(eventLog)
	}

	// Verify we only keep maxEvents
	events := svc.GetRecentEvents()
	if len(events) != svc.maxEvents {
		t.Errorf("Expected %d events, got %d", svc.maxEvents, len(events))
	}

	// Verify we kept the most recent ones
	if events[len(events)-1].Event.DstPort != 149 {
		t.Errorf("Expected last event port to be 149, got %d", events[len(events)-1].Event.DstPort)
	}
}
