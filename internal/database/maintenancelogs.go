package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetMaintenanceEditLogsAll(ctx context.Context) ([]models.MaintenanceEditLog, error) {
	start := time.Now()

	cursor, err := dbc.Collections.MaintenanceEditLogs.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to close cursor",
				"runtime", time.Since(start),
				"error", err,
			)
		}
	}()

	var logs []models.MaintenanceEditLog
	err = cursor.All(ctx, &logs)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved maintenance edit logs",
		"runtime", time.Since(start),
		"count", len(logs),
	)

	return logs, nil
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
			slog.ErrorContext(ctx, "failed to close cursor",
				"runtime", time.Since(start),
				"error", err,
			)
		}
	}()

	var logs []models.MaintenanceEditLog
	if err = cursor.All(ctx, &logs); err != nil {
		slog.ErrorContext(ctx, "failed to decode maintenance edit logs",
			"runtime", time.Since(start),
			"maintenance_id", maintenanceID,
			"error", err,
		)
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved maintenance edit logs",
		"runtime", time.Since(start),
		"maintenance_id", maintenanceID,
		"count", len(logs),
	)

	return logs, nil
}

func (dbc *Database) AddMaintenanceEditLog(ctx context.Context, entry *models.MaintenanceEditLog) error {
	start := time.Now()
	_, err := dbc.Collections.MaintenanceEditLogs.InsertOne(ctx, entry)
	if err != nil {
		slog.ErrorContext(ctx, "failed to insert maintenance edit log",
			"runtime", time.Since(start),
			"maintenance_id", entry.MaintenanceID,
			"error", err,
		)
	}

	slog.DebugContext(ctx, "inserted maintenance edit log",
		"runtime", time.Since(start),
		"maintenance_id", entry.MaintenanceID,
	)

	return err
}
