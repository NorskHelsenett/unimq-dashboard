package models

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

type MaintenanceStatus string

const (
	MaintenanceStatusScheduled  MaintenanceStatus = "scheduled"
	MaintenanceStatusInProgress MaintenanceStatus = "in_progress"
	MaintenanceStatusDone       MaintenanceStatus = "done"
	MaintenanceStatusSkipped    MaintenanceStatus = "skipped"
	MaintenanceStatusUnknown    MaintenanceStatus = "unknown"
)

func ParseMaintenanceStatus(s string) MaintenanceStatus {
	switch s {
	case "scheduled":
		return MaintenanceStatusScheduled
	case "in_progress":
		return MaintenanceStatusInProgress
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
		MaintenanceStatusInProgress,
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

	return slices.Contains(statuses, MaintenanceStatus(s))
}

// PostMaintenanceEntry is the model for creating a new maintenance entry
// The start and end time must follow the format "2006-01-02 15:04:05"
type PostMaintenanceEntry struct {
	Description string `json:"description" bson:"description" example:"maintenance for server upgrade"`
	Start       string `json:"start" bson:"start" example:"2024-06-01 10:00:00"`
	End         string `json:"end" bson:"end" example:"2024-06-01 12:00:00"`
}

func (p *PostMaintenanceEntry) ToMaintenanceEntry() (*MaintenanceEntry, error) {

	start, err := time.ParseInLocation(timeStampLayout, p.Start, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}

	end, err := time.ParseInLocation(timeStampLayout, p.End, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	return &MaintenanceEntry{
		ID:          uuid.New().String(),
		Description: p.Description,
		Start:       start,
		End:         end,
		Status:      MaintenanceStatusScheduled,
		Notified:    false,
	}, nil
}

type MaintenanceEntry struct {
	ID           string            `json:"id" bson:"_id" example:""`
	Description  string            `json:"description" bson:"description" example:"maintenance for server upgrade"`
	Start        time.Time         `json:"start" bson:"start" example:"2024-06-01 10:00:00"`
	End          time.Time         `json:"end" bson:"end" example:"2024-06-01 12:00:00"`
	Status       MaintenanceStatus `json:"status" bson:"status" example:"-"`
	Notified     bool              `json:"notified" bson:"notified"`
	UpdatedBy    string            `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	UpdatedAt    string            `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	UpdateReason string            `json:"update_reason,omitempty" bson:"update_reason,omitempty"`
}

const timeStampLayout = "2006-01-02 15:04:05"

func (e *MaintenanceEntry) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID           string `json:"id"`
		Description  string `json:"description"`
		Start        string `json:"start"`
		End          string `json:"end"`
		Status       string `json:"status"`
		Notified     bool   `json:"notified"`
		UpdatedBy    string `json:"updated_by"`
		UpdatedAt    string `json:"updated_at"`
		UpdateReason string `json:"update_reason"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	e.Start, err = time.Parse(timeStampLayout, aux.Start)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	e.End, err = time.Parse(timeStampLayout, aux.End)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	e.Description = aux.Description
	ok := IsValidMaintenanceStatus(aux.Status)
	if !ok {
		return fmt.Errorf("invalid maintenance status: %s, expected any of %v", aux.Status, GetMaintenanceStatusAllString())
	}
	e.Notified = false
	e.ID = aux.ID
	e.UpdatedBy = aux.UpdatedBy
	e.UpdatedAt = aux.UpdatedAt
	e.UpdateReason = aux.UpdateReason

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

type PatchMaintenanceEntry struct {
	Description string `json:"description"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Reason      string `json:"reason"`
	UpdatedBy   string `json:"updated_by"`
}

func (p *PatchMaintenanceEntry) Validate() error {
	if p.Description == "" {
		return fmt.Errorf("description is required")
	}
	if p.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if p.UpdatedBy == "" {
		return fmt.Errorf("updated_by is required")
	}
	if _, err := time.Parse(timeStampLayout, p.Start); err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}
	if _, err := time.Parse(timeStampLayout, p.End); err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}
	return nil
}

type MaintenanceEditLog struct {
	ID            string    `json:"id" bson:"_id"`
	MaintenanceID string    `json:"maintenance_id" bson:"maintenance_id"`
	Description   string    `json:"description" bson:"description"`
	Start         time.Time `json:"start" bson:"start"`
	End           time.Time `json:"end" bson:"end"`
	Reason        string    `json:"reason" bson:"reason"`
	UpdatedBy     string    `json:"updated_by" bson:"updated_by"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

func NewMaintenaceEditLog(maintenanceID string, description string, start time.Time, end time.Time, reason string, updatedBy string) *MaintenanceEditLog {
	return &MaintenanceEditLog{
		ID:            uuid.New().String(),
		MaintenanceID: maintenanceID,
		Description:   description,
		Start:         start,
		End:           end,
		Reason:        reason,
		UpdatedBy:     updatedBy,
		UpdatedAt:     time.Now(),
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

// Update status model for maintenance entry
//
//	@Description	Update the status of a maintenance entry (scheduled, done, skipped)
type UpdateMaintenance struct {
	Status MaintenanceStatus `json:"status"` // Enums - scheduled, done, skipped
}

func (u *UpdateMaintenance) UnmarshalJSON(data []byte) error {
	var aux struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if !IsValidMaintenanceStatus(aux.Status) {
		return fmt.Errorf("invalid maintenance status: %s", aux.Status)
	}

	u.Status = ParseMaintenanceStatus(aux.Status)
	return nil
}
