package models

import (
	"time"

	"github.com/google/uuid"
)

type AlarmEntry struct {
	AlarmID string     `bson:"_id"`
	Entries []LogEntry `bson:"entries"`
}

type LogEvent string

const (
	LogEventFired    LogEvent = "fired"
	LogEventResolved LogEvent = "resolved"
)

type LogEntry struct {
	ID        string    `json:"id" bson:"id"`
	Timestamp time.Time `json:"ts"`
	Event     LogEvent  `json:"event"`
	Value     *float64  `json:"value,omitempty"`
	Threshold float64   `json:"threshold"`
	AlarmType AlarmType `json:"alarm_type"`
}

func NewLogEntry(event LogEvent, value *float64, threshold float64, alarmType AlarmType) LogEntry {
	return LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Event:     event,
		Value:     value,
		Threshold: threshold,
		AlarmType: alarmType,
	}
}
