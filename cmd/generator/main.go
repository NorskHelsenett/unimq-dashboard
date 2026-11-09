package main

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/mongodb"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/services/storeservice"
)

var (
	aclEntries = []*models.ACL{
		{
			Group: "group.name1@local",
			Permissions: []models.Scope{
				models.ScopeAdmin,
			},
			VhostIDs: []string{"*"},
		},
		{
			Group: "acl-admin-group",
			Permissions: []models.Scope{
				models.ScopeAdmin,
			},
			VhostIDs: []string{"/"},
		},
		{
			Group: "acl-testers",
			Permissions: []models.Scope{
				models.ScopeRead,
				models.ScopeWrite,
			},
			VhostIDs: []string{"*"},
		},
	}

	alarmsSeed = []*models.AlarmRule{
		{
			ID:        "",
			QueueName: "",
			Name:      "High queue message size",
			Threshold: 50.0,
			Type:      models.AlarmTypeQueueMessages,
			Enabled:   true,
			Status:    models.AlarmStatusActive,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High queue size",
			Threshold: 15.0,
			Type:      models.AlarmTypeQueueSize,
			Enabled:   true,
			Status:    models.AlarmStatusActive,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(15.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High channel count",
			Threshold: 10.0,
			Type:      models.AlarmTypeChannels,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High connection count",
			Threshold: 10.0,
			Type:      models.AlarmTypeConnections,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
	}

	maintenanceEntries = []*models.MaintenanceEntry{
		{
			ID:          "e45957ef-b817-414e-95e8-c4ea89c4ad3e",
			Description: "Test maintenance entry 1",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusScheduled,
		},
		{
			ID:          "1fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 2",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusInProgress,
		},
		{
			ID:          "2fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 3",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusDone,
		},
	}
)

// This file is used to generate test data for the different stores. It is not used in the actual application.
func main() {

	logger.SetupLogger()

	config := config.NewConfig()
	if err := config.Load(); err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	err := config.CheckURLs()
	if err != nil {
		slog.Error("failed to validate URLs", "error", err)
		return
	}

	db, err := mongodb.NewDatabase(
		mongodb.WithDatabase(config.MongoDBDatabase),
		mongodb.WithHost(config.MongoDBHost),
		mongodb.WithPort(config.MongoDBPort),
		mongodb.WithUsername(config.MongoDBUsername),
		mongodb.WithPassword(config.MongoDBPassword),
	)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	err = db.InitCollections()
	if err != nil {
		slog.Error("failed to initialize database collections", "error", err)
		return
	}

	store, err := storeservice.NewStoreService(db)
	if err != nil {
		slog.Error("failed to create store service", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queueMapping := map[string][]string{
		"/":          {},
		"unimq":      {},
		"unimq-test": {},
	}

	vhosts := []string{}
	for vhost := range queueMapping {
		vhosts = append(vhosts, vhost)
	}

	rmq, err := rabbitmq.NewRMQClient(
		rabbitmq.WithRMQHost(config.RabbitMQHost),
		rabbitmq.WithRMQPort(config.RabbitMQPort),
		rabbitmq.WithRMQUsername(config.RabbitMQUsername),
		rabbitmq.WithRMQPassword(config.RabbitMQPassword),
		rabbitmq.WithRMQContext(ctx),
	)
	if err != nil {
		slog.Error("failed to create RabbitMQ client", "error", err)
		return
	}

	queueMapping, err = rmq.Seed(ctx, vhosts)
	if err != nil {
		slog.Error("failed to seed RabbitMQ", "error", err)
		return
	}

	slog.Info("RabbitMQ seeded successfully")

	err = db.Seed(ctx, queueMapping)
	if err != nil {
		slog.Error("failed to seed database", "error", err)
		return
	}
	slog.Info("Database seeded successfully")

	store.SeedACLs(ctx, aclEntries)
	slog.Info("ACLs seeded successfully")

	for _, vhost := range vhosts {

		queues := queueMapping[vhost]
		alarmSeedsCount := len(alarmsSeed)
		alarms := make([]*models.AlarmRule, 0, alarmSeedsCount)

		for _, queue := range queues {
			alarms = slices.Clone(alarmsSeed)
			for i := range alarms {
				alarms[i].ID = uuid.New().String()
				alarms[i].QueueName = queue
			}
		}

		err := store.SeedAlarms(ctx, alarms)
		if err != nil {
			slog.Error("failed to seed alarms", "error", err)
			return
		}
	}

	slog.Info("alarms seeded successfully")

	err = store.SeedMaintenace(ctx, maintenanceEntries)
	if err != nil {
		slog.Error("failed to seed maintenance entries", "error", err)
		return
	}
	slog.Info("maintenance entries seeded successfully")

	for _, entry := range maintenanceEntries {
		logEntries := make([]*models.MaintenanceEditLog, 0, 3)
		logEntries = append(logEntries, models.NewMaintenaceEditLog(entry.ID, "Test maintenance log entry 1", time.Now(), time.Now().Add(2*time.Hour), "Initial creation", "user1"))
		logEntries = append(logEntries, models.NewMaintenaceEditLog(entry.ID, "Test maintenance log entry 2", time.Now(), time.Now().Add(2*time.Hour), "Updated description", "user2"))
		logEntries = append(logEntries, models.NewMaintenaceEditLog(entry.ID, "Test maintenance log entry 3", time.Now(), time.Now().Add(2*time.Hour), "Updated start and end times", "user3"))
		err = store.SeedMaintenanceLogs(ctx, logEntries)
		if err != nil {
			slog.Error("failed to seed maintenance logs", "error", err)
			return
		}
	}

	slog.Info("maintenance logs seeded successfully")
}
