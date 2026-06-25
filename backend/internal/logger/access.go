package logger

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

// AccessEntry represents a single proxied request log entry.
type AccessEntry struct {
	Time       string  `json:"time"`
	Level      string  `json:"level"`
	ClientIP   string  `json:"client_ip"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Query      string  `json:"query"`
	Status     int     `json:"status"`
	DurationMS float64 `json:"duration_ms"`
	BytesSent  int64   `json:"bytes_sent"`
	Backend    string  `json:"backend"`
	BackendURL string  `json:"backend_url"`
	UserAgent  string  `json:"user_agent"`
	RequestID  string  `json:"request_id"`
}

// AccessLogger writes structured JSON access logs and broadcasts to SSE subscribers.
type AccessLogger struct {
	logger  *slog.Logger
	mu      sync.RWMutex
	subs    map[chan AccessEntry]struct{}
}

// NewAccessLogger creates a new access logger with JSON output.
func NewAccessLogger() *AccessLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return &AccessLogger{
		logger: slog.New(handler),
		subs:   make(map[chan AccessEntry]struct{}),
	}
}

// Log writes an access log entry and notifies SSE subscribers.
func (a *AccessLogger) Log(entry AccessEntry) {
	entry.Time = time.Now().UTC().Format(time.RFC3339)
	entry.Level = "INFO"

	a.logger.Info("access",
		"time", entry.Time,
		"client_ip", entry.ClientIP,
		"method", entry.Method,
		"path", entry.Path,
		"query", entry.Query,
		"status", entry.Status,
		"duration_ms", entry.DurationMS,
		"bytes_sent", entry.BytesSent,
		"backend", entry.Backend,
		"backend_url", entry.BackendURL,
		"user_agent", entry.UserAgent,
		"request_id", entry.RequestID,
	)

	a.mu.RLock()
	defer a.mu.RUnlock()
	for ch := range a.subs {
		select {
		case ch <- entry:
		default:
		}
	}
}

// Subscribe returns a channel that receives new log entries.
func (a *AccessLogger) Subscribe() chan AccessEntry {
	ch := make(chan AccessEntry, 100)
	a.mu.Lock()
	a.subs[ch] = struct{}{}
	a.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (a *AccessLogger) Unsubscribe(ch chan AccessEntry) {
	a.mu.Lock()
	delete(a.subs, ch)
	a.mu.Unlock()
	close(ch)
}

// MarshalEntry serializes an entry to JSON for SSE.
func MarshalEntry(entry AccessEntry) ([]byte, error) {
	return json.Marshal(entry)
}
