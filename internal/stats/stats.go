package stats

import (
	"sync"
	"time"
)

// ConnectionStat represents a tracked network connection stat.
type ConnectionStat struct {
	Timestamp   time.Time
	Application string
	Protocol    string
	Direction   string
	BytesSent   int64
	BytesRecv   int64
	Action      string // allow or deny
}

// Collector handles collection and retrieval of traffic stats.
type Collector struct {
	mu    sync.RWMutex
	stats []ConnectionStat
}

var defaultCollector = &Collector{
	stats: make([]ConnectionStat, 0, 1000),
}

// NewCollector creates a new stats collector.
func NewCollector() *Collector {
	return &Collector{
		stats: make([]ConnectionStat, 0, 1000),
	}
}

// Record records a connection stat.
func Record(stat ConnectionStat) {
	defaultCollector.Record(stat)
}

// Record adds a stat to the collector.
func (c *Collector) Record(stat ConnectionStat) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stat.Timestamp = time.Now()
	c.stats = append(c.stats, stat)

	// Keep last 10000 entries
	if len(c.stats) > 10000 {
		c.stats = c.stats[len(c.stats)-10000:]
	}
}

// Filter represents filtering criteria for stats.
type Filter struct {
	Application string
	Protocol    string
	Direction   string
	Action      string
	Since       time.Time
	Until       time.Time
}

// Query retrieves stats matching the filter.
func Query(filter Filter) []ConnectionStat {
	return defaultCollector.Query(filter)
}

// Query retrieves stats matching the filter.
func (c *Collector) Query(filter Filter) []ConnectionStat {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []ConnectionStat
	for _, stat := range c.stats {
		if filter.Application != "" && stat.Application != filter.Application {
			continue
		}
		if filter.Protocol != "" && stat.Protocol != filter.Protocol {
			continue
		}
		if filter.Direction != "" && stat.Direction != filter.Direction {
			continue
		}
		if filter.Action != "" && stat.Action != filter.Action {
			continue
		}
		if !filter.Since.IsZero() && stat.Timestamp.Before(filter.Since) {
			continue
		}
		if !filter.Until.IsZero() && stat.Timestamp.After(filter.Until) {
			continue
		}
		result = append(result, stat)
	}
	return result
}

// Snapshot returns summary statistics for this collector.
func (c *Collector) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := map[string]int64{
		"total_connections":   int64(len(c.stats)),
		"total_bytes_sent":    0,
		"total_bytes_recv":    0,
		"connections_allowed": 0,
		"connections_denied":  0,
	}

	for _, stat := range c.stats {
		result["total_bytes_sent"] += stat.BytesSent
		result["total_bytes_recv"] += stat.BytesRecv
		if stat.Action == "allow" {
			result["connections_allowed"]++
		} else if stat.Action == "deny" {
			result["connections_denied"]++
		}
	}

	return result
}

// Snapshot returns summary statistics from the default collector.
func Snapshot() (map[string]int64, error) {
	return defaultCollector.Snapshot(), nil
}

// Clear clears all collected statistics.
func Clear() {
	defaultCollector.Clear()
}

// Clear removes all stats from the collector.
func (c *Collector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats = make([]ConnectionStat, 0, 1000)
}

// GetTopApplications returns the top N applications by data transferred.
func GetTopApplications(n int) []struct {
	App        string
	BytesSent  int64
	BytesRecv  int64
	TotalBytes int64
} {
	defaultCollector.mu.RLock()
	defer defaultCollector.mu.RUnlock()

	// Aggregate by application
	appStats := make(map[string]*struct {
		App        string
		BytesSent  int64
		BytesRecv  int64
		TotalBytes int64
	})

	for _, stat := range defaultCollector.stats {
		if stat.Application == "" {
			continue
		}
		if _, exists := appStats[stat.Application]; !exists {
			appStats[stat.Application] = &struct {
				App        string
				BytesSent  int64
				BytesRecv  int64
				TotalBytes int64
			}{App: stat.Application}
		}
		appStats[stat.Application].BytesSent += stat.BytesSent
		appStats[stat.Application].BytesRecv += stat.BytesRecv
		appStats[stat.Application].TotalBytes += stat.BytesSent + stat.BytesRecv
	}

	// Convert to slice
	result := make([]struct {
		App        string
		BytesSent  int64
		BytesRecv  int64
		TotalBytes int64
	}, 0, len(appStats))

	for _, v := range appStats {
		result = append(result, *v)
	}

	// Sort by TotalBytes (simple bubble sort for small n)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].TotalBytes > result[i].TotalBytes {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Return top N
	if n > len(result) {
		n = len(result)
	}
	return result[:n]
}
