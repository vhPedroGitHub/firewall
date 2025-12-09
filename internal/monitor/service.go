package monitor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vhPedroGitHub/firewall/internal/logging"
	"github.com/vhPedroGitHub/firewall/internal/rules"
	"github.com/vhPedroGitHub/firewall/internal/stats"
)

// ConnectionEventLog represents a logged connection event with decision.
type ConnectionEventLog struct {
	Event     ConnectionEvent
	Decision  string // "allowed", "denied", "cancelled"
	Timestamp time.Time
	RuleName  string // Name of the rule that was applied or created
}

// ProcessTraffic tracks network traffic for a process.
type ProcessTraffic struct {
	AppPath       string
	BytesReceived int64
	BytesSent     int64
	Connections   int
	LastSeen      time.Time
}

// Service manages connection monitoring and user prompts.
type Service struct {
	monitor         Monitor
	handler         *DefaultHandler
	store           rules.Store
	stats           *stats.Collector
	running         bool
	promptsEnabled  bool
	eventsMu        sync.RWMutex
	recentEvts      []ConnectionEventLog
	maxEvents       int
	processesMu     sync.RWMutex
	activeProcesses map[string]ConnectionEvent // AppPath -> latest event
	trafficMu       sync.RWMutex
	processTraffic  map[string]*ProcessTraffic // AppPath -> traffic stats
}

// NewService creates a new monitoring service.
func NewService(store rules.Store) (*Service, error) {
	monitor, err := New()
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	handler := NewDefaultHandler(store)

	return &Service{
		monitor:         monitor,
		handler:         handler,
		store:           store,
		stats:           stats.NewCollector(),
		promptsEnabled:  true, // Enabled by default
		maxEvents:       100,  // Keep last 100 events
		recentEvts:      make([]ConnectionEventLog, 0, 100),
		activeProcesses: make(map[string]ConnectionEvent),
		processTraffic:  make(map[string]*ProcessTraffic),
	}, nil
}

// Start begins monitoring connections and handling prompts.
func (s *Service) Start() error {
	if s.running {
		return fmt.Errorf("service already running")
	}

	events, err := s.monitor.Start()
	if err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	s.running = true

	// Process events in background
	go s.processEvents(events)

	logging.LogEvent("info", "monitor_started", "Connection monitoring started", nil)
	return nil
}

// Stop stops the monitoring service.
func (s *Service) Stop() error {
	if !s.running {
		return fmt.Errorf("service not running")
	}

	if err := s.monitor.Stop(); err != nil {
		return fmt.Errorf("failed to stop monitor: %w", err)
	}

	s.running = false
	logging.LogEvent("info", "monitor_stopped", "Connection monitoring stopped", nil)
	return nil
}

// processEvents handles incoming connection events.
func (s *Service) processEvents(events <-chan ConnectionEvent) {
	for event := range events {
		// Track active process
		s.processesMu.Lock()
		s.activeProcesses[event.AppPath] = event
		s.processesMu.Unlock()

		// Track traffic (simplified - using estimated bytes)
		bytesTransferred := int64(1024) // placeholder estimate
		s.trackTraffic(event, bytesTransferred)

		// Record in stats module
		if event.AppPath != "" {
			action := "unknown"
			rulesList, _ := s.store.ListRules()
			for _, rule := range rulesList {
				if rule.Application == event.AppPath {
					action = rule.Action
					break
				}
			}

			s.stats.Record(stats.ConnectionStat{
				Timestamp:   time.Now(),
				Application: event.AppPath,
				Protocol:    event.Protocol,
				Direction:   event.Direction,
				BytesSent:   bytesTransferred / 2,
				BytesRecv:   bytesTransferred / 2,
				Action:      action,
			})
		}

		// Log the connection attempt
		logging.LogEvent("info", "connection_detected",
			fmt.Sprintf("Connection from %s (%s %s to %s:%d)",
				event.AppPath, event.Protocol, event.Direction, event.DstAddr, event.DstPort),
			nil)

		// Handle the connection (check rules and prompt if needed)
		decision, err := s.handler.HandleConnectionWithPrompts(event, s.promptsEnabled)
		if err != nil {
			log.Printf("Error handling connection: %v", err)
			logging.LogEvent("error", "connection_error",
				fmt.Sprintf("Failed to handle connection from %s: %v", event.AppPath, err),
				nil)
			continue
		}

		// Log the decision
		action := "denied"
		if decision == DecisionAllow {
			action = "allowed"
		} else if decision == DecisionCancel {
			action = "cancelled"
		}

		logging.LogEvent("info", "connection_"+action,
			fmt.Sprintf("Connection %s: %s (%s %s to %s:%d)",
				action, event.AppPath, event.Protocol, event.Direction, event.DstAddr, event.DstPort),
			nil)

		// Track the event
		ruleName := ""
		if decision != DecisionCancel {
			ruleName = fmt.Sprintf("auto_%s_%s_%d",
				sanitizeForRuleName(event.AppPath),
				event.Protocol,
				event.DstPort)
		}
		s.addEventLog(ConnectionEventLog{
			Event:     event,
			Decision:  action,
			Timestamp: time.Now(),
			RuleName:  ruleName,
		})

		// Save the decision as a rule for future connections
		if err := s.handler.SaveDecisionAsRule(event, decision); err != nil {
			log.Printf("Error saving rule: %v", err)
			logging.LogEvent("error", "rule_save_error",
				fmt.Sprintf("Failed to save rule for %s: %v", event.AppPath, err),
				nil)
		}
	}
}

// addEventLog adds an event to the recent events list.
func (s *Service) addEventLog(evt ConnectionEventLog) {
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()

	s.recentEvts = append(s.recentEvts, evt)

	// Keep only the most recent events
	if len(s.recentEvts) > s.maxEvents {
		s.recentEvts = s.recentEvts[len(s.recentEvts)-s.maxEvents:]
	}
}

// GetRecentEvents returns a copy of recent connection events.
func (s *Service) GetRecentEvents() []ConnectionEventLog {
	s.eventsMu.RLock()
	defer s.eventsMu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make([]ConnectionEventLog, len(s.recentEvts))
	copy(result, s.recentEvts)
	return result
}

// ClearEvents clears the event history.
func (s *Service) ClearEvents() {
	s.eventsMu.Lock()
	defer s.eventsMu.Unlock()
	s.recentEvts = make([]ConnectionEventLog, 0, s.maxEvents)
}

// IsRunning returns whether the service is currently running.
func (s *Service) IsRunning() bool {
	return s.running
}

// EnablePrompts enables automatic user prompts for unknown connections.
func (s *Service) EnablePrompts() {
	s.promptsEnabled = true
	logging.LogEvent("info", "prompts_enabled", "User prompts enabled", nil)
}

// DisablePrompts disables automatic user prompts for unknown connections.
func (s *Service) DisablePrompts() {
	s.promptsEnabled = false
	logging.LogEvent("info", "prompts_disabled", "User prompts disabled", nil)
}

// PromptsEnabled returns whether prompts are currently enabled.
func (s *Service) PromptsEnabled() bool {
	return s.promptsEnabled
}

// GetActiveProcesses returns a list of all processes attempting network connections.
func (s *Service) GetActiveProcesses() []ConnectionEvent {
	s.processesMu.RLock()
	defer s.processesMu.RUnlock()

	result := make([]ConnectionEvent, 0, len(s.activeProcesses))
	for _, evt := range s.activeProcesses {
		result = append(result, evt)
	}
	return result
}

// ClearActiveProcesses clears the active processes list.
func (s *Service) ClearActiveProcesses() {
	s.processesMu.Lock()
	defer s.processesMu.Unlock()
	s.activeProcesses = make(map[string]ConnectionEvent)
}

// trackTraffic updates traffic statistics for a connection.
// This is a simplified implementation - real traffic tracking would require
// packet capture or integration with network statistics APIs.
func (s *Service) trackTraffic(event ConnectionEvent, bytesTransferred int64) {
	s.trafficMu.Lock()
	defer s.trafficMu.Unlock()

	traffic, exists := s.processTraffic[event.AppPath]
	if !exists {
		traffic = &ProcessTraffic{
			AppPath:  event.AppPath,
			LastSeen: time.Now(),
		}
		s.processTraffic[event.AppPath] = traffic
	}

	traffic.Connections++
	traffic.LastSeen = time.Now()

	// Estimate bytes based on direction
	// In production, this would come from actual packet capture
	estimatedBytes := bytesTransferred
	if estimatedBytes == 0 {
		estimatedBytes = 1024 // Default estimate
	}

	if event.Direction == "outbound" {
		traffic.BytesSent += estimatedBytes
	} else {
		traffic.BytesReceived += estimatedBytes
	}
}

// GetProcessTraffic returns traffic statistics for all monitored processes.
func (s *Service) GetProcessTraffic() []ProcessTraffic {
	s.trafficMu.RLock()
	defer s.trafficMu.RUnlock()

	result := make([]ProcessTraffic, 0, len(s.processTraffic))
	for _, traffic := range s.processTraffic {
		result = append(result, *traffic)
	}
	return result
}

// ClearProcessTraffic clears the traffic statistics.
func (s *Service) ClearProcessTraffic() {
	s.trafficMu.Lock()
	defer s.trafficMu.Unlock()
	s.processTraffic = make(map[string]*ProcessTraffic)
}
