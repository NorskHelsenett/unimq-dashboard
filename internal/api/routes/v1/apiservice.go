package v1

import (
	"github.com/go-chi/chi/v5"
	api "github.com/sisneve/rabbitmq-dashboard/internal/api/v1"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupUtilityRoutes(r chi.Router, apiservice *api.APIService, dex *dex.DexClient, rmqHandler *rmq.RMQHandler) {

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/healthz", apiservice.HealthzHandler)
	r.Get("/readyz", apiservice.ReadyzHandler(rmqHandler, dex))
}

func SetupInternalRoutes(r chi.Router, apiservice *api.APIService) {

	r.Route("/maintenance", func(r chi.Router) {
		r.Get("/", apiservice.GetMaintenanceHandler)
		r.Get("/{maintenance-id}", apiservice.GetMaintenanceEntryHandler)
		r.Post("/", apiservice.AddMaintenanceHandler)
		r.Patch("/{maintenance-id}", apiservice.PatchMaintenanceHandler)
		r.Put("/{maintenance-id}", apiservice.UpdateMaintenanceStatusHandler)
		r.Delete("/{maintenance-id}", apiservice.DeleteMaintenanceHandler)
		r.Get("/{maintenance-id}/logs", apiservice.GetMaintenanceEditLogsHandler)
	})

	r.Route("/alarms", func(r chi.Router) {
		r.Get("/", apiservice.GetAlarmHistoryAllHandler)
		r.Get("/{rule-id}", apiservice.GetAlarmHistoryHandler)
	})

	r.Get("/status", apiservice.GetCheckerStatusHandler)
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", apiservice.GetNotificationsHandler)
		r.Route("/{vhost-name}", func(r chi.Router) {
			r.Get("/", apiservice.GetNotificationsVhostHandler)
			r.Delete("/", apiservice.DeleteNotificationsHandler)
			r.Post("/recipients", apiservice.AddNotificationsRecipientHandler)
			r.Get("/recipients/{recipient-id}", apiservice.GetNotificationsRecipientHandler)
			r.Delete("/recipients/{recipient-id}", apiservice.DeleteNotificationsRecipientHandler)

			r.Post("/rules", apiservice.AddNotificationsRuleHandler)
			r.Get("/rules/{rule-id}", apiservice.GetNotificationRuleHandler)
			r.Post("/rules/{rule-id}", apiservice.UpdateNotificationsRuleHandler)
			r.Post("/rules/{rule-id}/toggle", apiservice.ToggleNotificationsRuleHandler)
			r.Post("/rules/{rule-id}/test", apiservice.TestNotificationsRuleHandler)
			r.Delete("/rules/{rule-id}", apiservice.DeleteNotificationsRuleHandler)
		})
	})

}
