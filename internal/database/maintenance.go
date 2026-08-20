package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetMaintenanceAll(ctx context.Context, filter bson.D) ([]models.MaintenanceEntry, error) {
	start := time.Now()
	cursor, err := dbc.Collections.Maintenance.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to close cursor", "runtime", time.Since(start), "error", err)
		}
	}()

	var maintenance []models.MaintenanceEntry
	err = cursor.All(ctx, &maintenance)
	if err != nil {
		slog.ErrorContext(ctx, "failed to decode maintenance", "runtime", time.Since(start), "error", err)
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved maintenance", "runtime", time.Since(start), "count", len(maintenance))
	return maintenance, nil
}

func (dbc *Database) GetMaintenanceScheduled(ctx context.Context) ([]models.MaintenanceEntry, error) {
	return dbc.GetMaintenanceAll(ctx, bson.D{
		bson.E{
			Key: statusKey,
			Value: bson.M{
				"$in": []models.MaintenanceStatus{
					models.MaintenanceStatusScheduled,
					models.MaintenanceStatusInProgress,
				},
			},
		},
	})
}

func (dbc *Database) GetMaintenanceHistory(ctx context.Context) ([]models.MaintenanceEntry, error) {
	return dbc.GetMaintenanceAll(ctx, bson.D{
		bson.E{
			Key: statusKey,
			Value: bson.M{
				"$in": []models.MaintenanceStatus{
					models.MaintenanceStatusDone,
					models.MaintenanceStatusSkipped,
				},
			},
		},
	})
}

func (dbc *Database) AdvanceMaintenanceStatuses(ctx context.Context) (int64, error) {
	start := time.Now()
	now := time.Now().UTC()
	var total int64

	// scheduled → in_progress when start has passed but end has not
	result, err := dbc.Collections.Maintenance.UpdateMany(ctx,
		bson.M{
			statusKey: models.MaintenanceStatusScheduled,
			"start":   bson.M{"$lte": now},
			"end":     bson.M{"$gt": now},
		},
		bson.M{set: bson.M{statusKey: models.MaintenanceStatusInProgress}},
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to advance maintenance statuses to in_progress", "runtime", time.Since(start), "error", err)
		return 0, err
	}
	if result.ModifiedCount > 0 {
		slog.InfoContext(ctx, "advanced maintenance statuses to in_progress", "runtime", time.Since(start), "count", result.ModifiedCount)
	}
	total += result.ModifiedCount

	// scheduled or in_progress → done when end has passed
	result, err = dbc.Collections.Maintenance.UpdateMany(ctx,
		bson.M{
			statusKey: bson.M{"$in": []models.MaintenanceStatus{
				models.MaintenanceStatusScheduled,
				models.MaintenanceStatusInProgress,
			}},
			"end": bson.M{"$lte": now},
		},
		bson.M{set: bson.M{statusKey: models.MaintenanceStatusDone}},
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to advance maintenance statuses to done", "runtime", time.Since(start), "error", err)
		return total, err
	}
	if result.ModifiedCount > 0 {
		slog.InfoContext(ctx, "advanced maintenance statuses to done", "runtime", time.Since(start), "count", result.ModifiedCount)
	}
	total += result.ModifiedCount

	return total, nil
}

func (dbc *Database) SetMaintenanceEntryStatus(ctx context.Context, entryID string, status models.MaintenanceStatus) error {
	start := time.Now()

	filter := map[string]any{id: entryID}
	update := map[string]any{
		set: map[string]any{
			statusKey: status,
		},
	}
	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), id, entryID, statusKey, status, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), id, entryID, statusKey, status)
	}

	return err
}

func (dbc *Database) SetMaintenanceEntryNotified(ctx context.Context, entryID string, notified bool) error {
	start := time.Now()

	filter := map[string]any{id: entryID}
	update := map[string]any{
		set: map[string]any{
			"notified": notified,
		},
	}
	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), id, entryID, "notified", notified, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), id, entryID, "notified", notified)
	}

	return err
}

func (dbc *Database) GetMaintenanceEntry(ctx context.Context, entryID string) (*models.MaintenanceEntry, error) {
	start := time.Now()

	var entry models.MaintenanceEntry
	err := dbc.Collections.Maintenance.FindOne(ctx, map[string]any{id: entryID}).Decode(&entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve maintenance", "runtime", time.Since(start), id, entryID, "error", err)
		return nil, err
	}
	slog.DebugContext(ctx, "retrieved maintenance", "runtime", time.Since(start), id, entryID)

	return &entry, err
}

func (dbc *Database) AddMaintenanceEntry(ctx context.Context, entry *models.MaintenanceEntry) error {
	start := time.Now()

	_, err := dbc.Collections.Maintenance.InsertOne(ctx, entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create maintenance", "runtime", time.Since(start), id, entry.ID, "error", err)
	} else {
		slog.DebugContext(ctx, "created maintenance", "runtime", time.Since(start), id, entry.ID)
	}

	return err
}

func (dbc *Database) UpdateMaintenanceEntry(ctx context.Context, entryID string, status string) error {
	start := time.Now()

	filter := map[string]any{id: entryID}
	update := map[string]any{
		set: map[string]any{
			statusKey: status,
		},
	}

	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), id, entryID, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), id, entryID)
	}

	return err
}

func (dbc *Database) PatchMaintenanceEntry(ctx context.Context, entryID string, description string, start time.Time, end time.Time, reason string, updatedBy string) error {
	tstart := time.Now()

	updatedAt := tstart.Format("2006-01-02 15:04:05")
	filter := map[string]any{id: entryID}
	update := map[string]any{
		set: map[string]any{
			"description":   description,
			"start":         start,
			"end":           end,
			"update_reason": reason,
			"updated_by":    updatedBy,
			"updated_at":    updatedAt,
		},
	}

	result, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to patch maintenance", "runtime", time.Since(tstart), id, entryID, "error", err)
		return err
	}
	if result.MatchedCount == 0 {
		return ErrMaintenanceNotFound
	}
	slog.DebugContext(ctx, "patched maintenance", "runtime", time.Since(tstart), id, entryID)
	return nil
}

var (
	ErrMaintenanceNotFound = fmt.Errorf("maintenance entry not found")
)

func (dbc *Database) AddMaintenanceEditLog(ctx context.Context, entry *models.MaintenanceEditLog) error {
	start := time.Now()
	_, err := dbc.Collections.MaintenanceEditLogs.InsertOne(ctx, entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to insert maintenance edit log", "runtime", time.Since(start), "maintenance_id", entry.MaintenanceID, "error", err)
	} else {
		slog.DebugContext(ctx, "inserted maintenance edit log", "runtime", time.Since(start), "maintenance_id", entry.MaintenanceID)
	}
	return err
}

func (dbc *Database) GetMaintenanceEditLogs(ctx context.Context, maintenanceID string) ([]models.MaintenanceEditLog, error) {
	start := time.Now()
	opts := bson.D{{Key: "maintenance_id", Value: maintenanceID}}
	cursor, err := dbc.Collections.MaintenanceEditLogs.Find(ctx, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close cursor", "runtime", time.Since(start), "error", err)
		}
	}()

	var logs []models.MaintenanceEditLog
	if err = cursor.All(ctx, &logs); err != nil {
		slog.ErrorContext(ctx, "failed to decode maintenance edit logs", "runtime", time.Since(start), "maintenance_id", maintenanceID, "error", err)
		return nil, err
	}
	slog.DebugContext(ctx, "retrieved maintenance edit logs", "runtime", time.Since(start), "maintenance_id", maintenanceID, "count", len(logs))
	return logs, nil
}

func (dbc *Database) DeleteMaintenanceEntry(ctx context.Context, entryID string) error {
	start := time.Now()

	status, err := dbc.Collections.Maintenance.DeleteOne(ctx, map[string]any{id: entryID})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete maintenance", "runtime", time.Since(start), id, entryID, "error", err)
		return err
	}
	if status.DeletedCount == 0 {
		slog.ErrorContext(ctx, "no maintenance entry found to delete", "runtime", time.Since(start), id, entryID)
		return fmt.Errorf("%w, with id: %s", ErrMaintenanceNotFound, entryID)
	}

	slog.DebugContext(ctx, "deleted maintenance", "runtime", time.Since(start), id, entryID)
	return nil
}
