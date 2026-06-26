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
			Key:   "status",
			Value: models.MaintenanceStatusScheduled,
		},
	},
	)
}

func (dbc *Database) GetMaintenanceHistory(ctx context.Context) ([]models.MaintenanceEntry, error) {
	return dbc.GetMaintenanceAll(ctx, bson.D{
		bson.E{
			Key: "status",
			Value: bson.M{
				"$ne": models.MaintenanceStatusScheduled,
			},
		},
	},
	)

}

func (dbc *Database) SetMaintenanceEntryStatus(ctx context.Context, id string, status models.MaintenanceStatus) error {
	start := time.Now()

	filter := map[string]any{"_id": id}
	update := map[string]any{
		"$set": map[string]any{
			"status": status,
		},
	}
	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), "_id", id, "status", status, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), "_id", id, "status", status)
	}

	return err
}

func (dbc *Database) SetMaintenanceEntryNotified(ctx context.Context, id string, notified bool) error {
	start := time.Now()

	filter := map[string]any{"_id": id}
	update := map[string]any{
		"$set": map[string]any{
			"notified": notified,
		},
	}
	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), "_id", id, "notified", notified, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), "_id", id, "notified", notified)
	}

	return err
}

func (dbc *Database) GetMaintenanceEntry(ctx context.Context, id string) (*models.MaintenanceEntry, error) {
	start := time.Now()

	var entry models.MaintenanceEntry
	err := dbc.Collections.Maintenance.FindOne(ctx, map[string]any{"_id": id}).Decode(&entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve maintenance", "runtime", time.Since(start), "id", id, "error", err)
		return nil, err
	}
	slog.DebugContext(ctx, "retrieved maintenance", "runtime", time.Since(start), "_id", id)

	return &entry, err
}

func (dbc *Database) AddMaintenanceEntry(ctx context.Context, entry *models.MaintenanceEntry) error {
	start := time.Now()

	_, err := dbc.Collections.Maintenance.InsertOne(ctx, entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create maintenance", "runtime", time.Since(start), "_id", entry.ID, "error", err)
	} else {
		slog.DebugContext(ctx, "created maintenance", "runtime", time.Since(start), "_id", entry.ID)
	}

	return err
}

func (dbc *Database) UpdateMaintenanceEntry(ctx context.Context, id string, status string) error {
	start := time.Now()

	filter := map[string]any{"_id": id}
	update := map[string]any{
		"$set": map[string]any{
			"status": status,
		},
	}

	_, err := dbc.Collections.Maintenance.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update maintenance", "runtime", time.Since(start), "_id", id, "error", err)
	} else {
		slog.DebugContext(ctx, "updated maintenance", "runtime", time.Since(start), "_id", id)
	}

	return err
}

var (
	ErrMaintenanceNotFound = fmt.Errorf("maintenance entry not found")
)

func (dbc *Database) DeleteMaintenanceEntry(ctx context.Context, id string) error {
	start := time.Now()

	status, err := dbc.Collections.Maintenance.DeleteOne(ctx, map[string]any{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete maintenance", "runtime", time.Since(start), "_id", id, "error", err)
		return err
	}
	if status.DeletedCount == 0 {
		slog.ErrorContext(ctx, "no maintenance entry found to delete", "runtime", time.Since(start), "_id", id)
		return fmt.Errorf("%w, with id: %s", ErrMaintenanceNotFound, id)
	}

	slog.DebugContext(ctx, "deleted maintenance", "runtime", time.Since(start), "_id", id)
	return nil
}
