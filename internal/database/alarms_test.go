package database_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func TestAlarms(t *testing.T) {

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

	value := 42.0
	err = db.AddAlarm(ctx, &models.AlarmEntry{
		AlarmID: "test-alarm",
		Entries: []models.LogEntry{
			{
				Timestamp: time.Now(),
				Event:     models.LogEvent("Test alarm triggered"),
				Value:     &value,
				Threshold: 40.0,
			},
		},
	})
	if err != nil {
		slog.Error("failed to add test alarm", "error", err)
	}

	alarm, err := db.GetAlarm(ctx, "test-alarm")
	if err != nil {
		slog.Error("failed to get test alarm", "error", err)
	} else {
		slog.Info("retrieved alarm", "id", alarm.AlarmID, "entries", alarm.Entries)
	}

	alarms, err := db.GetAlarmsAll(ctx)
	if err != nil {
		slog.Error("failed to get alarms", "error", err)
	} else {
		for _, alarm := range alarms {
			slog.Info("alarm entry", "id", alarm.AlarmID, "entries", alarm.Entries)
		}
	}

	err = db.DeleteAlarm(ctx, "test-alarm")
	if err != nil {
		slog.Error("failed to delete test alarm", "error", err)
	}

}
