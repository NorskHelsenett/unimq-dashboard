package models

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
)

type MaintenanceStatus string

const (
	MaintenanceStatusScheduled MaintenanceStatus = "scheduled"
	MaintenanceStatusDone      MaintenanceStatus = "done"
	MaintenanceStatusSkipped   MaintenanceStatus = "skipped"
	MaintenanceStatusUnknown   MaintenanceStatus = "unknown"
)

func ParseMaintenanceStatus(s string) MaintenanceStatus {
	switch s {
	case "scheduled":
		return MaintenanceStatusScheduled
	case "done":
		return MaintenanceStatusDone
	case "skipped":
		return MaintenanceStatusSkipped
	default:
		return MaintenanceStatusUnknown
	}
}

func GetMaintenanceStatusAll() []MaintenanceStatus {
	return []MaintenanceStatus{
		MaintenanceStatusScheduled,
		MaintenanceStatusDone,
		MaintenanceStatusSkipped,
	}
}

func GetMaintenanceStatusAllString() []string {
	status := GetMaintenanceStatusAll()
	out := make([]string, len(status))
	for i, s := range status {
		out[i] = string(s)
	}
	return out
}

func IsValidMaintenanceStatus(s string) bool {
	statuses := GetMaintenanceStatusAll()

	if slices.Contains(statuses, MaintenanceStatus(s)) {
		return true
	}
	return false
}

type MaintenanceEntry struct {
	ID          string            `json:"id" bson:"_id"`
	Description string            `json:"description" bson:"description"`
	Start       time.Time         `json:"start" bson:"start"`
	End         time.Time         `json:"end" bson:"end"`
	Status      MaintenanceStatus `json:"status" bson:"status"`
	Notified    bool              `json:"notified" bson:"notified"`
}

func (e *MaintenanceEntry) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID          string            `json:"id"`
		Description string            `json:"description"`
		Start       string            `json:"start"`
		End         string            `json:"end"`
		Status      MaintenanceStatus `json:"status"`
		Notified    bool              `json:"notified"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	e.Start, err = time.Parse(time.RFC3339, aux.Start)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	e.End, err = time.Parse(time.RFC3339, aux.End)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	return nil
}

func NewMaintenanceEntry(description string, start time.Time, end time.Time) *MaintenanceEntry {
	return &MaintenanceEntry{
		Description: description,
		Start:       start,
		End:         end,
		Status:      MaintenanceStatusScheduled,
		Notified:    false,
	}
}

type MaintenanceAdminResponse struct {
	Entries []MaintenanceEntry
}

func NewMaintenanceAdminResponse(entries []MaintenanceEntry) *MaintenanceAdminResponse {
	return &MaintenanceAdminResponse{
		Entries: entries,
	}
}

type MaintenanceResponse struct {
	Scheduled []MaintenanceEntry
	History   []MaintenanceEntry
}

func NewMaintenanceResponse(scheduled []MaintenanceEntry, history []MaintenanceEntry) *MaintenanceResponse {
	return &MaintenanceResponse{
		Scheduled: scheduled,
		History:   history,
	}
}
