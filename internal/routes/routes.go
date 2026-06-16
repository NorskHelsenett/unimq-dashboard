package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sisneve/rabbitmq-dashboard/internal/api"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func SetupRoutes(ctx context.Context, config *config.Config, db *database.Database, rmq *rabbitmq.RMQClient) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)

	prom, err := prometheus.NewPromClient(config.PrometheusHost, "v1", "", "", config.PrometheusPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	limits := &models.Limits{
		MaxChannels:    config.RabbitMQChannelLimit,
		MaxConnections: config.RabbitMQConnectionLimit,
		MaxQueues:      config.RabbitMQQueueLimit,
	}

	apiservice, err := api.NewAPIService(
		api.WithContext(ctx),
		api.WithRabbitMQClient(rmq),
		api.WithPromClient(prom),
		api.WithDatabase(db),
		api.WithRMQLimits(limits),
	)

	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	// // Serve the callback page so oidc-client-ts can complete the OIDC flow client-side.
	// http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "web/templates/callback.html")
	// })

	// TODO: group into v1, v2, etc. as needed for API versioning and better organization
	// TODO: correct the methods for these routes (e.g. POST for add/delete operations)

	// HTML routes
	r.Get("/", apiservice.IndexHandler)
	r.Get("/queue", apiservice.QueueHandler)
	r.Get("/maintenance", apiservice.GetMaintenanceHandler)
	r.Get("/maintenance/admin", apiservice.GetMaintenanceAdminHandler)
	r.Post("/maintenance/add", apiservice.PostMaintenanceAddHandler)
	r.Post("/maintenance/status", apiservice.PostMaintenanceStatusHandler)
	r.Post("/maintenance/delete", apiservice.PostMaintenanceDeleteHandler)
	r.Get("/notifications", apiservice.GetNotificationsHandler)

	// TODO: where are recipients listed?
	r.Post("/notifications/recipients/add", apiservice.PostNotificationsAddRecipientHandler)
	r.Post("/notifications/recipients/delete", apiservice.PostNotificationsDeleteRecipientHandler)
	r.Post("/notifications/rules/add", apiservice.PostNotificationsAddRuleHandler)
	r.Post("/notifications/rules/delete", apiservice.PostNotificationsDeleteRuleHandler)
	r.Post("/notifications/rules/update", apiservice.PostNotificationsUpdateRuleHandler)

	// TODO: Where are rules listed?
	r.Post("/notifications/rules/toggle", apiservice.PostNotificationsToggleRuleHandler)
	r.Post("/notifications/rules/message", apiservice.PostNotificationsUpdateMessageHandler) // TODO: requires rule ID and vhost query params, likely a Get request
	r.Post("/notifications/rules/test", apiservice.PostNotificationsTestHandler)
	r.Get("/notifications/rules/logs", apiservice.NotificationsLogsHandler) // TODO: Likely a Get request, requires an id query param, unsure which id (rule id? vhost? log id?)
	r.Get("/notifications/rule", apiservice.NotificationsRuleHandler)       // TODO: Likely a Get request, requires rule ID and vhost query params
	r.Get("/profile", apiservice.GetProfileHandler)                         // TODO: requires vhost query param

	// API routes
	r.Get("/api/queues", apiservice.GetQueuesHandler) // TODO: Fix required vhost query param
	r.Get("/api/cluster", apiservice.GetClusterHandler)

	// Logs every route implicitly or explicitly defined above with its method, path, and number of middlewares.
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("route info", "method", method, "route", route, "middlewares", len(r.Middlewares()))
		return nil
	})

	return r, nil

}
