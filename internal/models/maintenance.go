package models

import (
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
	Status      MaintenanceStatus `json:"status" bson:"status"` // "scheduled", "done", "skipped"
	Notified    bool              `json:"notified" bson:"notified"`
}
