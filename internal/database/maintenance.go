package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetMaintenanceAll(ctx context.Context) ([]models.MaintenanceEntry, error) {
	start := time.Now()
	cursor, err := dbc.Collections.Maintenance.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var maintenance []models.MaintenanceEntry
	err = cursor.All(ctx, &maintenance)
	if err != nil {
		return nil, err
	}

	slog.Info("retrieved maintenance", "runtime", time.Since(start), "count", len(maintenance))
	return maintenance, nil
}

// TODO: Use Collection object to cursor.
func (dbc *Database) GetMaintenanceEntry(ctx context.Context, id string) (*models.MaintenanceEntry, error) {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("maintenance")

	var entry models.MaintenanceEntry
	err := collection.FindOne(ctx, map[string]any{"id": id}).Decode(&entry)
	slog.Info("retrieved maintenance", "runtime", time.Since(start), "id", id)
	return &entry, err
}

func (dbc *Database) AddMaintenanceEntry(ctx context.Context, entry *models.MaintenanceEntry) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("maintenance")

	_, err := collection.InsertOne(ctx, entry)
	slog.Info("created maintenance", "runtime", time.Since(start), "id", entry.ID)
	return err
}

func (dbc *Database) UpdateMaintenanceEntry(ctx context.Context, id string, status string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("maintenance")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"id": id},
		map[string]any{
			"$set": map[string]any{
				"status": status,
			},
		},
	)
	slog.Info("updated maintenance", "runtime", time.Since(start), "id", id)
	return err
}

func (dbc *Database) DeleteMaintenanceEntry(ctx context.Context, id string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("maintenance")

	_, err := collection.DeleteOne(ctx, map[string]any{"id": id})
	slog.Info("deleted maintenance", "runtime", time.Since(start), "id", id)
	return err
}
