package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaintenanceEntry_UnmarshalJSON_Valid(t *testing.T) {
	data := []byte(`{
		"id": "abc-123",
		"description": "server upgrade",
		"start": "2024-06-01T10:00:00Z",
		"end": "2024-06-01T12:00:00Z",
		"status": "scheduled",
		"notified": true,
		"updated_by": "jane@example.com",
		"updated_at": "2024-05-01T08:00:00Z",
		"update_reason": "created"
	}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.NoError(t, err)

	assert.Equal(t, "abc-123", entry.ID)
	assert.Equal(t, "server upgrade", entry.Description)
	assert.Equal(t, time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), entry.Start)
	assert.Equal(t, time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), entry.End)
	assert.Equal(t, models.MaintenanceStatusScheduled, entry.Status)
	assert.True(t, entry.Notified)
	assert.Equal(t, "jane@example.com", entry.UpdatedBy)
	assert.Equal(t, time.Date(2024, 5, 1, 8, 0, 0, 0, time.UTC), entry.UpdatedAt)
	assert.Equal(t, "created", entry.UpdateReason)
}

func TestMaintenanceEntry_UnmarshalJSON_DefaultsUpdatedAtToNow(t *testing.T) {
	data := []byte(`{
		"start": "2024-06-01T10:00:00Z",
		"end": "2024-06-01T12:00:00Z",
		"status": "scheduled"
	}`)

	before := time.Now()
	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	after := time.Now()
	require.NoError(t, err)

	assert.True(t, !entry.UpdatedAt.Before(before) && !entry.UpdatedAt.After(after),
		"expected UpdatedAt to default to approximately now, got %v", entry.UpdatedAt)
}

func TestMaintenanceEntry_UnmarshalJSON_MissingStart(t *testing.T) {
	data := []byte(`{"end": "2024-06-01T12:00:00Z", "status": "scheduled"}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start is required")
}

func TestMaintenanceEntry_UnmarshalJSON_MissingEnd(t *testing.T) {
	data := []byte(`{"start": "2024-06-01T10:00:00Z", "status": "scheduled"}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end is required")
}

func TestMaintenanceEntry_UnmarshalJSON_InvalidStartFormat(t *testing.T) {
	data := []byte(`{"start": "2024-06-01 10:00:00", "end": "2024-06-01T12:00:00Z", "status": "scheduled"}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start time format")
}

func TestMaintenanceEntry_UnmarshalJSON_InvalidEndFormat(t *testing.T) {
	data := []byte(`{"start": "2024-06-01T10:00:00Z", "end": "2024-06-01 12:00:00", "status": "scheduled"}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid end time format")
}

func TestMaintenanceEntry_UnmarshalJSON_InvalidUpdatedAtFormat(t *testing.T) {
	data := []byte(`{
		"start": "2024-06-01T10:00:00Z",
		"end": "2024-06-01T12:00:00Z",
		"status": "scheduled",
		"updated_at": "not-a-date"
	}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid updated_at time format")
}

func TestMaintenanceEntry_UnmarshalJSON_InvalidStatus(t *testing.T) {
	data := []byte(`{
		"start": "2024-06-01T10:00:00Z",
		"end": "2024-06-01T12:00:00Z",
		"status": "not-a-real-status"
	}`)

	var entry models.MaintenanceEntry
	err := json.Unmarshal(data, &entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid maintenance status")
}

func TestMaintenanceEntry_UnmarshalJSON_MalformedJSON(t *testing.T) {
	var entry models.MaintenanceEntry
	err := json.Unmarshal([]byte(`{"start":`), &entry)
	assert.Error(t, err)
}

func TestUpdateMaintenance_UnmarshalJSON_Valid(t *testing.T) {
	cases := []struct {
		status   string
		expected models.MaintenanceStatus
	}{
		{"scheduled", models.MaintenanceStatusScheduled},
		{"in_progress", models.MaintenanceStatusInProgress},
		{"done", models.MaintenanceStatusDone},
		{"skipped", models.MaintenanceStatusSkipped},
	}

	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			var u models.UpdateMaintenance
			err := json.Unmarshal([]byte(`{"status":"`+tc.status+`"}`), &u)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, u.Status)
		})
	}
}

func TestUpdateMaintenance_UnmarshalJSON_InvalidStatus(t *testing.T) {
	var u models.UpdateMaintenance
	err := json.Unmarshal([]byte(`{"status":"not-a-real-status"}`), &u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid maintenance status")
}

func TestUpdateMaintenance_UnmarshalJSON_MalformedJSON(t *testing.T) {
	var u models.UpdateMaintenance
	err := json.Unmarshal([]byte(`{"status":`), &u)
	assert.Error(t, err)
}
