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

	r.Get("/api/v1/vhosts", apiservice.VhostsHandler)
	r.Get("/api/v1/vhosts/{vhost}", apiservice.VhostHandler)
	r.Get("/api/v1/vhosts/{vhost}/metrics", apiservice.MetricHandler)
	r.Get("/api/v1/vhosts/{vhost}/queues", apiservice.GetQueuesHandler)
	r.Get("/api/v1/vhosts/{vhost}/queues/{queue}", apiservice.GetQueuesByNameHandler)

	r.Get("/api/v1/vhosts/{vhost}/notifications", apiservice.GetNotificationsHandler)
	r.Post("/api/v1/vhosts/{vhost}/notifications/recipients", apiservice.AddNotificationsRecipientHandler)
	r.Delete("/api/v1/vhosts/{vhost}/notifications/recipients/{recipient}", apiservice.DeleteNotificationsRecipientHandler)

	r.Post("/api/v1/vhosts/{vhost}/notifications/rules", apiservice.AddNotificationsRuleHandler)
	r.Post("/api/v1/vhosts/{vhost}/notifications/rules/{rule}", apiservice.PostNotificationsUpdateRuleHandler)
	r.Post("/api/v1/vhosts/{vhost}/notifications/rules/{rule}/toggle", apiservice.ToggleNotificationsRuleHandler)
	r.Post("/api/v1/vhosts/{vhost}/notifications/rules/{rule}/message", apiservice.PostNotificationsUpdateMessageHandler)
	r.Post("/api/v1/vhosts/{vhost}/notifications/rules/{rule}/test", apiservice.PostNotificationsTestHandler)
	r.Delete("/api/v1/vhosts/{vhost}/notifications/rules/{rule}", apiservice.DeleteNotificationsRuleHandler)

	r.Get("/api/v1/vhosts/{vhost}/notifications/logs", apiservice.NotificationsLogsHandler)

	r.Get("/api/v1/maintenance", apiservice.GetMaintenanceHandler)
	r.Get("/api/v1/maintenance/admin", apiservice.GetMaintenanceAdminHandler)
	r.Post("/api/v1/maintenance", apiservice.AddMaintenanceHandler)
	r.Post("/api/v1/maintenance/{maintenance}", apiservice.UpdateMaintenanceStatusHandler)
	r.Delete("/api/v1/maintenance/{maintenance}", apiservice.DeleteMaintenanceHandler)

	r.Get("/api/v1/cluster", apiservice.GetClusterHandler)

	// Logs every route implicitly or explicitly defined above with its method, path, and number of middlewares.
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("route info", "method", method, "route", route, "middlewares", len(r.Middlewares()))
		return nil
	})

	return r, nil

}
