package store

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const maxLogEntries = 100

type LogEvent string

const (
	LogEventFired    LogEvent = "fired"
	LogEventResolved LogEvent = "resolved"
)

type LogEntry struct {
	Timestamp time.Time `json:"ts"`
	Event     LogEvent  `json:"event"`
	Value     *float64  `json:"value,omitempty"`
	Threshold float64   `json:"threshold"`
}

type LogStore struct {
	mu      sync.RWMutex
	path    string
	Entries map[string][]LogEntry `json:"entries"`
}

func NewLogStore(path string) (*LogStore, error) {
	ls := &LogStore{
		path:    path,
		Entries: make(map[string][]LogEntry),
	}
	if err := ls.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ls, nil
}

func (ls *LogStore) load() error {
	data, err := os.ReadFile(ls.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, ls)
}

func (ls *LogStore) save() error {
	data, err := json.MarshalIndent(ls, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ls.path, data, 0644)
}

func (ls *LogStore) Append(alarmID string, entry LogEntry) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	entries := ls.Entries[alarmID]
	entries = append(entries, entry)
	if len(entries) > maxLogEntries {
		entries = entries[len(entries)-maxLogEntries:]
	}
	ls.Entries[alarmID] = entries
	return ls.save()
}

func (ls *LogStore) Get(alarmID string) []LogEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	src := ls.Entries[alarmID]
	out := make([]LogEntry, len(src))
	copy(out, src)
	// Newest first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
