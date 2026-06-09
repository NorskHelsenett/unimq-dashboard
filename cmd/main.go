package main

import (
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/maintenance"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify/store"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
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

	templating.MaintStore, err = maintenance.NewStore("data/maintenance.json")
	if err != nil {
		log.Fatalf("could not load maintenance store: %v", err)
	}

	templating.NotifyStore, err = store.NewStore("data/notifications.json")
	if err != nil {
		log.Fatalf("could not load notification store: %v", err)
	}

	templating.LogStore, err = store.NewLogStore("data/alarm_logs.json")
	if err != nil {
		log.Fatalf("could not load alarm log store: %v", err)
	}

	checker := notify.NewChecker(
		notify.WithConfig(config),
		notify.WithStore(templating.NotifyStore),
		notify.WithLogStore(templating.LogStore),
		notify.WithMaintStore(templating.MaintStore),
		notify.WithInterval(60*time.Second),
	)
	checker.StartChecker()

	if err := mime.AddExtensionType(".js", "application/javascript"); err != nil {
		log.Fatalf("failed to register MIME type: %v", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	// Serve the callback page so oidc-client-ts can complete the OIDC flow client-side.
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/callback.html")
	})

	routes := routes.SetupRoutes(config)

	slog.Info("starting RabbitMQ Dashboard", "URL", config.BaseURL, "port", config.BasePort)
	err = http.ListenAndServe(fmt.Sprintf("%v:%d", config.BaseURL, config.BasePort), routes)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
