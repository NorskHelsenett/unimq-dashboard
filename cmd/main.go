package main

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
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

	err := config.CheckURLs()
	if err != nil {
		slog.Error("failed to validate URLs", "error", err)
		return
	}

	checker := notify.NewChecker(
		notify.WithInterval(60 * time.Second),
	)
	checker.StartChecker()

	if err := mime.AddExtensionType(".js", "application/javascript"); err != nil {
		slog.Error("failed to register MIME type", "error", err)
		return
	}

	ctx := context.Background()
	routes, err := routes.SetupRoutes(ctx, config)
	if err != nil {
		slog.Error("failed to set up routes", "error", err)
		return
	}

	slog.Info("starting RabbitMQ Dashboard", "URL", config.BaseURL, "port", config.BasePort)
	err = http.ListenAndServe(fmt.Sprintf("%v:%d", config.BaseURL, config.BasePort), routes)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		return
	}

}
