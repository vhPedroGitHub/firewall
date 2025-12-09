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

// Snapshot returns summary statistics.
func Snapshot() (map[string]int64, error) {
	defaultCollector.mu.RLock()
	defer defaultCollector.mu.RUnlock()

	result := map[string]int64{
		"total_connections": int64(len(defaultCollector.stats)),
		"total_bytes_sent":  0,
		"total_bytes_recv":  0,
	}

	for _, stat := range defaultCollector.stats {
		result["total_bytes_sent"] += stat.BytesSent
		result["total_bytes_recv"] += stat.BytesRecv
	}

	return result, nil
}
