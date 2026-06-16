package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
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

	uri := database.BuildURI(config.MongoDBHost, config.MongoDBUsername, config.MongoDBPassword, config.MongoDBPort)
	db, err := database.NewDatabase(uri, config.MongoDBDatabase)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.Seed(ctx)
	if err != nil {
		slog.Error("failed to seed database", "error", err)
		return
	}

	slog.Info("database seeded successfully")
}
