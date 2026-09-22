package routes

import (
	"github.com/go-chi/chi/v5"
	api "github.com/sisneve/rabbitmq-dashboard/internal/api/v1"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupUnprotectedRoutes(r chi.Router, apiservice *api.APIService, dex *dex.DexClient) {

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/healthz", apiservice.HealthzHandler)
	r.Get("/readyz", apiservice.ReadyzHandler(dex))
}

func SetupProtectedRoutes(r chi.Router, apiservice *api.APIService, rmqhandler *rmq.RMQHandler) {

	r.Route("/v1", func(r chi.Router) {
		r.Route("/vhosts", func(r chi.Router) {
			r.Get("/", apiservice.VhostsHandler)
			r.Get("/{vhost}", apiservice.VhostHandler)
			r.Get("/{vhost}/metrics", rmqhandler.MetricHandler)

			r.Route("/{vhost}/queues", func(r chi.Router) {
				r.Get("/", rmqhandler.GetQueuesHandler)
				r.Get("/{queue}", rmqhandler.GetQueuesByNameHandler)
			})

		})
		r.Route("/maintenance", func(r chi.Router) {
			r.Get("/", apiservice.GetMaintenanceHandler)
			r.Get("/{maintenance}", apiservice.GetMaintenanceEntryHandler)
			r.Post("/", apiservice.AddMaintenanceHandler)
			r.Patch("/{maintenance}", apiservice.PatchMaintenanceHandler)
			r.Put("/{maintenance}", apiservice.UpdateMaintenanceStatusHandler)
			r.Delete("/{maintenance}", apiservice.DeleteMaintenanceHandler)
			r.Get("/{maintenance}/logs", apiservice.GetMaintenanceEditLogsHandler)
		})

		r.Route("/alarms", func(r chi.Router) {
			r.Get("/", apiservice.GetAlarmHistoryAllHandler)
			r.Get("/{rule-id}", apiservice.GetAlarmHistoryHandler)
		})

		r.Route("/rabbitmq", func(r chi.Router) {
			r.Get("/nodes", rmqhandler.GetRMQNodesHandler)
			r.Get("/vhostusage", rmqhandler.GetRMQVhostUsageHandler)
		})
		r.Get("/status", apiservice.GetCheckerStatusHandler)
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", apiservice.GetNotificationsHandler)
			r.Route("/{vhost}", func(r chi.Router) {
				r.Get("/", apiservice.GetNotificationsVhostHandler)
				r.Delete("/", apiservice.DeleteNotificationsHandler)
				r.Post("/recipients", apiservice.AddNotificationsRecipientHandler)
				r.Get("/recipients/{recipient}", apiservice.GetNotificationsRecipientHandler)
				r.Delete("/recipients/{recipient}", apiservice.DeleteNotificationsRecipientHandler)

				r.Post("/rules", apiservice.AddNotificationsRuleHandler)
				r.Get("/rules/{rule}", apiservice.GetNotificationRuleHandler)
				r.Post("/rules/{rule}", apiservice.UpdateNotificationsRuleHandler)
				r.Post("/rules/{rule}/toggle", apiservice.ToggleNotificationsRuleHandler)
				r.Post("/rules/{rule}/test", apiservice.TestNotificationsRuleHandler)
				r.Delete("/rules/{rule}", apiservice.DeleteNotificationsRuleHandler)
			})
		})
	})

}
