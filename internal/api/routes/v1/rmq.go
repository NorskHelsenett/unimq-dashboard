package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
)

func SetupRMQRoutes(r chi.Router, rmqhandler *rmq.RMQHandler) {

	r.Route("/vhosts", func(r chi.Router) {
		r.Get("/", rmqhandler.GetVhostsHandler)
		r.Get("/{vhost-name}", rmqhandler.GetVhostHandler)
		r.Get("/{vhost-name}/metrics", rmqhandler.MetricHandler)
		r.Get("/{vhost-name}/limits", rmqhandler.GetVhostLimitsHandler)
		r.Get("/{vhost-name}/usage", rmqhandler.GetRMQVhostUsageHandler)

		r.Route("/{vhost-name}/queues", func(r chi.Router) {
			r.Get("/", rmqhandler.GetQueuesHandler)
			r.Get("/{queue-id}", rmqhandler.GetQueuesByNameHandler)
		})

	})

	r.Route("/rabbitmq", func(r chi.Router) {
		r.Get("/", rmqhandler.GetRMQNodesHandler)
	})
}
