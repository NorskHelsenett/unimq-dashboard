package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

func SetupRoutes(config *config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)

	rmqclient := scraper.NewRestClient(
		fmt.Sprintf("%v:%d/api",
			config.RabbitMQHost,
			config.RabbitMQPort,
		),
		config.RabbitMQUsername,
		config.RabbitMQPassword,
		config.PrometheusHost,
		"v1",
		config.PrometheusPort,
	)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	// Serve the callback page so oidc-client-ts can complete the OIDC flow client-side.
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/callback.html")
	})

	// TODO: group into v1, v2, etc. as needed for API versioning and better organization
	// TODO: correct the methods for these routes (e.g. POST for add/delete operations)

	// HTML routes
	r.Get("/", rmqclient.IndexHandler)
	r.Get("/queue", rmqclient.QueueHandler)
	r.Get("/maintenance", rmqclient.GetMaintenanceHandler)
	r.Get("/maintenance/admin", rmqclient.GetMaintenanceAdminHandler)
	r.Post("/maintenance/add", rmqclient.MaintenanceAddHandler)
	r.Get("/maintenance/status", rmqclient.MaintenanceStatusHandler)
	r.Post("/maintenance/delete", rmqclient.MaintenanceDeleteHandler)
	r.Get("/notifications", rmqclient.GetNotificationsHandler)
	// TODO: where are recipients listed?
	r.Post("/notifications/recipients/add", rmqclient.PostNotificationsAddRecipientHandler)
	r.Post("/notifications/recipients/delete", rmqclient.PostNotificationsDeleteRecipientHandler)
	r.Post("/notifications/rules/add", rmqclient.PostNotificationsAddRuleHandler)
	r.Post("/notifications/rules/delete", rmqclient.PostNotificationsDeleteRuleHandler)
	r.Post("/notifications/rules/update", rmqclient.PostNotificationsUpdateRuleHandler)
	// TODO: Where are rules listed?
	r.Post("/notifications/rules/toggle", rmqclient.PostNotificationsToggleRuleHandler)
	r.Post("/notifications/rules/message", rmqclient.PostNotificationsUpdateMessageHandler) // TODO: requires rule ID and vhost query params, likely a Get request
	r.Post("/notifications/rules/test", rmqclient.PostNotificationsTestHandler)
	r.HandleFunc("/notifications/rules/logs", rmqclient.NotificationsLogsHandler) // TODO: Likely a Get request, requires an id query param, unsure which id (rule id? vhost? log id?)
	r.HandleFunc("/notifications/rule", rmqclient.NotificationsRuleHandler)       // TODO: Likely a Get request, requires rule ID and vhost query params
	r.Get("/profile", rmqclient.GetProfileHandler)                                // TODO: requires vhost query param

	// API routes
	r.Get("/api/queues", rmqclient.GetQueuesHandler) // TODO: Fix required vhost query param
	r.Get("/api/cluster", rmqclient.GetClusterHandler)

	// Logs every route implicitly or explicitly defined above with its method, path, and number of middlewares.
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("route info", "method", method, "route", route, "middlewares", len(r.Middlewares()))
		return nil
	})

	return r

}
