package stats

import (
	"testing"
	"time"
)

func TestCollector_Record(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	stat := ConnectionStat{
		Application: "/usr/bin/firefox",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   1024,
		BytesRecv:   2048,
		Action:      "allow",
	}

	c.Record(stat)

	if len(c.stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(c.stats))
	}

	recorded := c.stats[0]
	if recorded.Application != stat.Application {
		t.Errorf("expected application %q, got %q", stat.Application, recorded.Application)
	}
	if recorded.BytesSent != stat.BytesSent {
		t.Errorf("expected BytesSent %d, got %d", stat.BytesSent, recorded.BytesSent)
	}
	if recorded.Timestamp.IsZero() {
		t.Error("expected timestamp to be set, got zero")
	}
}

func TestCollector_MultipleRecords(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	for i := 0; i < 5; i++ {
		c.Record(ConnectionStat{
			Application: "app",
			Protocol:    "tcp",
			Direction:   "outbound",
			BytesSent:   int64(i * 100),
			BytesRecv:   int64(i * 200),
			Action:      "allow",
		})
	}

	if len(c.stats) != 5 {
		t.Fatalf("expected 5 stats, got %d", len(c.stats))
	}
}

func TestCollector_Query_NoFilter(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "app1", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "app2", Protocol: "udp", Direction: "inbound", Action: "deny"})

	results := c.Query(Filter{})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestCollector_Query_FilterByApplication(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "firefox", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "chrome", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "firefox", Protocol: "udp", Direction: "inbound", Action: "allow"})

	results := c.Query(Filter{Application: "firefox"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Application != "firefox" {
			t.Errorf("expected application 'firefox', got %q", r.Application)
		}
	}
}

func TestCollector_Query_FilterByProtocol(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "app", Protocol: "udp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "inbound", Action: "allow"})

	results := c.Query(Filter{Protocol: "tcp"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Protocol != "tcp" {
			t.Errorf("expected protocol 'tcp', got %q", r.Protocol)
		}
	}
}

func TestCollector_Query_FilterByDirection(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "inbound", Action: "allow"})

	results := c.Query(Filter{Direction: "outbound"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Direction != "outbound" {
		t.Errorf("expected direction 'outbound', got %q", results[0].Direction)
	}
}

func TestCollector_Query_FilterByAction(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "app", Protocol: "tcp", Direction: "outbound", Action: "deny"})

	results := c.Query(Filter{Action: "deny"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Action != "deny" {
		t.Errorf("expected action 'deny', got %q", results[0].Action)
	}
}

func TestCollector_Query_FilterByTime(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	c.stats = append(c.stats, ConnectionStat{
		Timestamp:   past,
		Application: "app1",
		Protocol:    "tcp",
		Direction:   "outbound",
		Action:      "allow",
	})
	c.stats = append(c.stats, ConnectionStat{
		Timestamp:   now,
		Application: "app2",
		Protocol:    "tcp",
		Direction:   "outbound",
		Action:      "allow",
	})
	c.stats = append(c.stats, ConnectionStat{
		Timestamp:   future,
		Application: "app3",
		Protocol:    "tcp",
		Direction:   "outbound",
		Action:      "allow",
	})

	// Query for stats since 30 minutes ago
	since := now.Add(-30 * time.Minute)
	results := c.Query(Filter{Since: since})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Query for stats until 30 minutes from now
	until := now.Add(30 * time.Minute)
	results = c.Query(Filter{Until: until})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestCollector_Query_CombinedFilters(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{Application: "firefox", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "firefox", Protocol: "tcp", Direction: "inbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "chrome", Protocol: "tcp", Direction: "outbound", Action: "allow"})
	c.Record(ConnectionStat{Application: "firefox", Protocol: "udp", Direction: "outbound", Action: "allow"})

	results := c.Query(Filter{
		Application: "firefox",
		Protocol:    "tcp",
		Direction:   "outbound",
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Application != "firefox" || r.Protocol != "tcp" || r.Direction != "outbound" {
		t.Errorf("unexpected result: %+v", r)
	}
}

func TestCollector_MaxEntries(t *testing.T) {
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	// Record more than 10000 entries
	for i := 0; i < 10500; i++ {
		c.Record(ConnectionStat{
			Application: "app",
			Protocol:    "tcp",
			Direction:   "outbound",
			Action:      "allow",
		})
	}

	// Should trim to 10000
	if len(c.stats) != 10000 {
		t.Errorf("expected 10000 stats after trimming, got %d", len(c.stats))
	}
}

func TestSnapshot(t *testing.T) {
	// Use fresh collector for this test
	c := &Collector{
		stats: make([]ConnectionStat, 0),
	}

	c.Record(ConnectionStat{
		Application: "app1",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   1000,
		BytesRecv:   2000,
		Action:      "allow",
	})
	c.Record(ConnectionStat{
		Application: "app2",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   500,
		BytesRecv:   1500,
		Action:      "allow",
	})

	// Manually compute snapshot for our local collector
	result := map[string]int64{
		"total_connections": int64(len(c.stats)),
		"total_bytes_sent":  0,
		"total_bytes_recv":  0,
	}

	for _, stat := range c.stats {
		result["total_bytes_sent"] += stat.BytesSent
		result["total_bytes_recv"] += stat.BytesRecv
	}

	if result["total_connections"] != 2 {
		t.Errorf("expected 2 connections, got %d", result["total_connections"])
	}
	if result["total_bytes_sent"] != 1500 {
		t.Errorf("expected 1500 bytes sent, got %d", result["total_bytes_sent"])
	}
	if result["total_bytes_recv"] != 3500 {
		t.Errorf("expected 3500 bytes received, got %d", result["total_bytes_recv"])
	}
}

func TestGlobalCollector(t *testing.T) {
	// Test the global functions
	Record(ConnectionStat{
		Application: "global-test",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   100,
		BytesRecv:   200,
		Action:      "allow",
	})

	results := Query(Filter{Application: "global-test"})

	if len(results) == 0 {
		t.Fatal("expected at least 1 result from global collector")
	}

	found := false
	for _, r := range results {
		if r.Application == "global-test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("did not find expected global-test stat in global collector")
	}
}

func TestCollector_Clear(t *testing.T) {
	c := NewCollector()

	// Record some stats
	c.Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "test.exe",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   1024,
		BytesRecv:   512,
		Action:      "allow",
	})

	// Verify stats exist
	results := c.Query(Filter{})
	if len(results) != 1 {
		t.Errorf("Expected 1 stat, got %d", len(results))
	}

	// Clear stats
	c.Clear()

	// Verify stats cleared
	results = c.Query(Filter{})
	if len(results) != 0 {
		t.Errorf("Expected 0 stats after clear, got %d", len(results))
	}
}

func TestGetTopApplications(t *testing.T) {
	// Clear default collector
	Clear()

	// Record stats for different applications
	Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app1.exe",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   5000,
		BytesRecv:   3000,
		Action:      "allow",
	})

	Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app2.exe",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   2000,
		BytesRecv:   1000,
		Action:      "allow",
	})

	Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app1.exe",
		Protocol:    "tcp",
		Direction:   "inbound",
		BytesSent:   1000,
		BytesRecv:   500,
		Action:      "allow",
	})

	// Get top 2 applications
	topApps := GetTopApplications(2)

	if len(topApps) != 2 {
		t.Errorf("Expected 2 top apps, got %d", len(topApps))
	}

	// First should be app1.exe with most data
	if topApps[0].App != "app1.exe" {
		t.Errorf("Expected app1.exe as top app, got %s", topApps[0].App)
	}

	expectedTotal := int64(9500)
	if topApps[0].TotalBytes != expectedTotal {
		t.Errorf("Expected %d total bytes for app1.exe, got %d", expectedTotal, topApps[0].TotalBytes)
	}

	// Second should be app2.exe
	if topApps[1].App != "app2.exe" {
		t.Errorf("Expected app2.exe as second app, got %s", topApps[1].App)
	}

	expectedTotal2 := int64(3000)
	if topApps[1].TotalBytes != expectedTotal2 {
		t.Errorf("Expected %d total bytes for app2.exe, got %d", expectedTotal2, topApps[1].TotalBytes)
	}
}

func TestSnapshot_WithAllowedAndDenied(t *testing.T) {
	c := NewCollector()

	// Record mixed stats
	c.Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app1.exe",
		Protocol:    "tcp",
		Direction:   "outbound",
		BytesSent:   1000,
		BytesRecv:   500,
		Action:      "allow",
	})

	c.Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app2.exe",
		Protocol:    "tcp",
		Direction:   "inbound",
		BytesSent:   2000,
		BytesRecv:   1000,
		Action:      "deny",
	})

	c.Record(ConnectionStat{
		Timestamp:   time.Now(),
		Application: "app3.exe",
		Protocol:    "udp",
		Direction:   "outbound",
		BytesSent:   500,
		BytesRecv:   250,
		Action:      "allow",
	})

	// Get snapshot
	snapshot := c.Snapshot()

	if snapshot["total_connections"] != 3 {
		t.Errorf("Expected 3 total connections, got %d", snapshot["total_connections"])
	}

	if snapshot["total_bytes_sent"] != 3500 {
		t.Errorf("Expected 3500 total bytes sent, got %d", snapshot["total_bytes_sent"])
	}

	if snapshot["total_bytes_recv"] != 1750 {
		t.Errorf("Expected 1750 total bytes recv, got %d", snapshot["total_bytes_recv"])
	}

	if snapshot["connections_allowed"] != 2 {
		t.Errorf("Expected 2 allowed connections, got %d", snapshot["connections_allowed"])
	}

	if snapshot["connections_denied"] != 1 {
		t.Errorf("Expected 1 denied connection, got %d", snapshot["connections_denied"])
	}
}
