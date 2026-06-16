package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes"
)

func main() {

	logger.SetupLogger()

	config := config.NewConfig()
	if err := config.Load(); err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	logger.UpdateLogLevel(slog.Level(config.LogLevel))

	err := config.CheckURLs()
	if err != nil {
		slog.Error("failed to validate URLs", "error", err)
		return
	}

	// if err := mime.AddExtensionType(".js", "application/javascript"); err != nil {
	// 	slog.Error("failed to register MIME type", "error", err)
	// 	return
	// }

	uri := database.BuildURI(config.MongoDBHost, config.MongoDBUsername, config.MongoDBPassword, config.MongoDBPort)
	db, err := database.NewDatabase(uri, config.MongoDBDatabase)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	ctx := context.Background()

	rmqurl := fmt.Sprintf("%v:%d/api", config.RabbitMQHost, config.RabbitMQPort)
	rmq, err := rabbitmq.NewRMQClient(ctx, rmqurl, config.RabbitMQUsername, config.RabbitMQPassword)
	if err != nil {
		slog.Error("failed to create RabbitMQ client", "error", err)
		return
	}

	routes, err := routes.SetupRoutes(ctx, config, db, rmq)
	if err != nil {
		slog.Error("failed to set up routes", "error", err)
		return
	}

	slog.Info("starting RabbitMQ Dashboard", "URL", config.BaseURL, "port", config.BasePort)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	wg.Go(func() {
		defer wg.Done()
		err = http.ListenAndServe(fmt.Sprintf("%v:%d", config.BaseURL, config.BasePort), routes)
		if err != nil {
			slog.Error("failed to start server", "error", err)
			return
		}
	})

	checker := notify.NewChecker(
		notify.WithDB(db),
		notify.WithRMQClient(rmq),
		notify.WithInterval(60*time.Second),
	)
	checker.StartChecker()
	wg.Wait()

}
