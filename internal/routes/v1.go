package routes

import (
	"github.com/go-chi/chi/v5"
	api "github.com/sisneve/rabbitmq-dashboard/internal/api/v1"
)

func SetupV1Routes(r chi.Router, apiservice *api.APIService) {

	r.Route("/v1", func(r chi.Router) {
		r.Route("/vhosts", func(r chi.Router) {
			r.Get("/", apiservice.VhostsHandler)
			r.Get("/{vhost}", apiservice.VhostHandler)
			r.Get("/{vhost}/metrics", apiservice.MetricHandler)

			r.Route("/{vhost}/queues", func(r chi.Router) {
				r.Get("/", apiservice.GetQueuesHandler)
				r.Get("/{queue}", apiservice.GetQueuesByNameHandler)
			})

		})
		r.Route("/maintenance", func(r chi.Router) {
			r.Get("/", apiservice.GetMaintenanceHandler)
			r.Get("/admin", apiservice.GetMaintenanceAdminHandler)
			r.Post("/", apiservice.AddMaintenanceHandler)
			r.Patch("/{maintenance}", apiservice.PatchMaintenanceHandler)
			r.Put("/{maintenance}", apiservice.UpdateMaintenanceStatusHandler)
			r.Delete("/{maintenance}", apiservice.DeleteMaintenanceHandler)
			r.Get("/{maintenance}/logs", apiservice.GetMaintenanceEditLogsHandler)
		})

		r.Route("/alarms", func(r chi.Router) {
			r.Get("/", apiservice.GetAlarmHistoryAllHandler)
			r.Get("/{vhost}", apiservice.GetAlarmHistoryHandler)
		})

		r.Get("/cluster", apiservice.GetClusterHandler)
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", apiservice.GetNotificationsHandler)
			r.Route("/{vhost}", func(r chi.Router) {
				r.Get("/", apiservice.GetNotificationsVhostHandler)
				r.Delete("/", apiservice.DeleteNotificationsHandler)
				r.Post("/recipients", apiservice.AddNotificationsRecipientHandler)
				r.Delete("/recipients/{recipient}", apiservice.DeleteNotificationsRecipientHandler)

				r.Post("/rules", apiservice.AddNotificationsRuleHandler)
				r.Post("/rules/{rule}", apiservice.UpdateNotificationsRuleHandler)
				r.Post("/rules/{rule}/toggle", apiservice.ToggleNotificationsRuleHandler)
				r.Post("/rules/{rule}/test", apiservice.TestNotificationsRuleHandler)
				r.Delete("/rules/{rule}", apiservice.DeleteNotificationsRuleHandler)
			})
		})
	})

}
