package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Event represents a firewall event entry.
type Event struct {
	Timestamp time.Time
	Level     string // info, warning, error
	Category  string // rule-add, rule-remove, connection-allow, connection-deny, etc.
	Message   string
	Details   map[string]interface{}
}

// Logger handles structured event logging.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	filepath string
}

var defaultLogger *Logger

// Init initializes the default logger with a file path.
func Init(filepath string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defaultLogger = &Logger{
		file:     f,
		filepath: filepath,
	}
	return nil
}

// Close closes the default logger.
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// LogEvent logs an event to the default logger.
func LogEvent(level, category, message string, details map[string]interface{}) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Log(level, category, message, details)
}

// Log writes an event to the log file as JSON.
func (l *Logger) Log(level, category, message string, details map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	event := Event{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Details:   details,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintln(l.file, string(data))
}

// ReadEvents reads events from a log file (simple implementation).
func ReadEvents(filepath string) ([]Event, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var events []Event
	// Each line is a JSON event
	lines := string(data)
	var current []byte
	for i := 0; i < len(lines); i++ {
		if lines[i] == '\n' {
			if len(current) > 0 {
				var e Event
				if err := json.Unmarshal(current, &e); err == nil {
					events = append(events, e)
				}
				current = nil
			}
		} else {
			current = append(current, lines[i])
		}
	}
	return events, nil
}
