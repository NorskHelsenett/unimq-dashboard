package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
)

func SetupRMQRoutes(r chi.Router, rmqhandler *rmq.RMQHandler) {

	r.Route("/vhosts", func(r chi.Router) {
		r.Get("/", rmqhandler.VhostsHandler)
		r.Get("/{vhost}", rmqhandler.VhostHandler)
		r.Get("/{vhost}/metrics", rmqhandler.MetricHandler)

		r.Route("/{vhost}/queues", func(r chi.Router) {
			r.Get("/", rmqhandler.GetQueuesHandler)
			r.Get("/{queue}", rmqhandler.GetQueuesByNameHandler)
		})

	})

	r.Route("/rabbitmq", func(r chi.Router) {
		r.Get("/nodes", rmqhandler.GetRMQNodesHandler)
		r.Get("/vhostusage", rmqhandler.GetRMQVhostUsageHandler)
	})
}
