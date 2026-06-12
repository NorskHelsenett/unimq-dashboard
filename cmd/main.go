package main

import (
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes"
)

func main() {

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
		log.Fatalf("failed to register MIME type: %v", err)
		return
	}

	routes, err := routes.SetupRoutes(config)
	if err != nil {
		log.Fatalf("failed to set up routes: %v", err)
		return
	}

	slog.Info("starting RabbitMQ Dashboard", "URL", config.BaseURL, "port", config.BasePort)
	err = http.ListenAndServe(fmt.Sprintf("%v:%d", config.BaseURL, config.BasePort), routes)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
		return
	}

}
