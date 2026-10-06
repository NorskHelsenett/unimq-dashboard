package models_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMaintenanceStart = "2024-06-01T10:00:00Z"
	testMaintenanceEnd   = "2024-06-01T12:00:00Z"
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

func TestPatchMaintenanceEntry_Validate(t *testing.T) {
	valid := func() *models.PatchMaintenanceEntry {
		return &models.PatchMaintenanceEntry{
			Description: "server upgrade",
			Start:       testMaintenanceStart,
			End:         testMaintenanceEnd,
			Reason:      "updated maintenance time",
		}
	}

	t.Run("valid", func(t *testing.T) {
		assert.NoError(t, valid().Validate())
	})

	t.Run("missing description", func(t *testing.T) {
		p := valid()
		p.Description = ""
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description is required")
	})

	t.Run("missing reason", func(t *testing.T) {
		p := valid()
		p.Reason = ""
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reason is required")
	})

	t.Run("invalid start format", func(t *testing.T) {
		p := valid()
		p.Start = "2024-06-01 10:00:00"
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid start time format")
	})

	t.Run("invalid end format", func(t *testing.T) {
		p := valid()
		p.End = "2024-06-01 12:00:00"
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid end time format")
	})

	t.Run("end before start", func(t *testing.T) {
		p := valid()
		p.Start = testMaintenanceEnd
		p.End = testMaintenanceStart
		err := p.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "end time must be after start time")
	})

	t.Run("end equal to start is accepted", func(t *testing.T) {
		p := valid()
		p.Start = testMaintenanceStart
		p.End = testMaintenanceStart
		assert.NoError(t, p.Validate(), "end.Before(start) is false when they are equal, so this currently passes validation")
	})
}

const testMaintenanceEmail = "jane@example.com"

func contextWithEmail() context.Context {
	return context.WithValue(context.Background(), httpsuite.ClaimsContextKey, map[string]any{"email": testMaintenanceEmail})
}

func TestPostMaintenanceEntry_ToMaintenanceEntry(t *testing.T) {
	valid := func() *models.PostMaintenanceEntry {
		return &models.PostMaintenanceEntry{
			Description: "server upgrade",
			Start:       testMaintenanceStart,
			End:         testMaintenanceEnd,
		}
	}

	t.Run("valid", func(t *testing.T) {
		ctx := contextWithEmail()
		before := time.Now()
		entry, err := valid().ToMaintenanceEntry(ctx)
		after := time.Now()
		require.NoError(t, err)

		assert.NotEmpty(t, entry.ID)
		assert.Equal(t, "server upgrade", entry.Description)
		assert.Equal(t, time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC), entry.Start)
		assert.Equal(t, time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC), entry.End)
		assert.Equal(t, models.MaintenanceStatusScheduled, entry.Status)
		assert.Equal(t, "jane@example.com", entry.UpdatedBy)
		assert.True(t, !entry.UpdatedAt.Before(before) && !entry.UpdatedAt.After(after))
		assert.Equal(t, "created", entry.UpdateReason)
		assert.False(t, entry.Notified)
	})

	t.Run("invalid start format", func(t *testing.T) {
		p := valid()
		p.Start = "2024-06-01 10:00:00"
		_, err := p.ToMaintenanceEntry(contextWithEmail())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid start time format")
	})

	t.Run("invalid end format", func(t *testing.T) {
		p := valid()
		p.End = "2024-06-01 12:00:00"
		_, err := p.ToMaintenanceEntry(contextWithEmail())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid end time format")
	})

	t.Run("end before start", func(t *testing.T) {
		p := valid()
		p.Start = testMaintenanceEnd
		p.End = testMaintenanceStart
		_, err := p.ToMaintenanceEntry(contextWithEmail())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "end time must be after start time")
	})

	t.Run("missing email in context", func(t *testing.T) {
		_, err := valid().ToMaintenanceEntry(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get email from context")
	})
}
