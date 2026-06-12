package models

import "time"

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
	Timestamp time.Time `json:"ts"`
	Event     LogEvent  `json:"event"`
	Value     *float64  `json:"value,omitempty"`
	Threshold float64   `json:"threshold"`
}
