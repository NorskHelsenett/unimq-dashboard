package database_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMaintenance(t *testing.T) {
	config := config.NewConfig()
	err := config.Load()
	if err != nil {
		panic(err)
	}

	err = config.CheckURLs()
	if err != nil {
		panic(err)
	}

	uri := database.BuildURI(config.MongoDBHost, config.MongoDBUsername, config.MongoDBPassword, config.MongoDBPort)
	db, err := database.NewDatabase(uri, config.MongoDBDatabase)
	if err != nil {
		panic(err)
	}

	err = db.InitCollections()
	if err != nil {
		panic(err)
	}

	ctx := t.Context()

	err = db.AddMaintenanceEntry(ctx, &models.MaintenanceEntry{
		ID:          "test-maintenance",
		Description: "Test maintenance entry",
		Start:       time.Now(),
		End:         time.Now().Add(2 * time.Hour),
		Status:      models.MaintenanceStatusScheduled,
	})
	if err != nil {
		slog.Error("failed to add test maintenance", "error", err)
	}

	entry, err := db.GetMaintenanceEntry(ctx, "test-maintenance")
	if err != nil {
		slog.Error("failed to get test entry", "error", err)
	} else {
		slog.Info("retrieved entry", "id", entry.ID)
	}

	entries, err := db.GetMaintenanceAll(ctx, bson.M{})
	if err != nil {
		slog.Error("failed to get test entries", "error", err)
	} else {
		slog.Info("entry", "entries", len(entries))
	}

	err = db.DeleteMaintenanceEntry(ctx, "test-maintenance")
	if err != nil {
		slog.Error("failed to delete test entry", "error", err)
	}

}
