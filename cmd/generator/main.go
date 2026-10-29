package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/mongodb"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/services/storeservice"
)

var aclEntries = []models.ACL{
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

	dbint := database.Store(db)
	store := storeservice.NewStoreService(dbint)

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

	store.SeedACLs(ctx, aclEntries)

	slog.Info("database seeded successfully")
}
