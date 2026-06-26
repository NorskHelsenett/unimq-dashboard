package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Queue Metrics for a specific queue in the specified virtual host
// @Description	Fetches time-series metrics for a specific RabbitMQ queue over a specified time range.
// @Tags			Queues
// @Produce		html
// @Param			vhost	query		string				true	"Virtual Host"
// @Param			name	query		string				true	"Queue Name"
// @Param			range	query		string				false	"Time Range (e.g., 1h, 24h, 7d)"	default(1h)
// @Success		200		{string}	string				"HTML page with queue metrics"
// @Failure		400		{object}	httpsuite.APIError	"Bad Request"
// @Failure		500		{object}	httpsuite.APIError	"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues [get]
func (rc *APIService) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	queues, err := rc.RMQClient.GetQueues()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queues", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queues", http.StatusOK, &queues)
}

// @Summary		Get All Queues to a specified virtual host
// @Description	Fetches details of all queues in a specified virtual host.
// @Tags			Queues
// @Produce		json
// @Param			vhost	query		string				true	"Virtual Host"
// @Param			queue	query		string				true	"Queue Name"
// @Success		200		{array}		models.QueueDetail	"List of queue details"
// @Failure		400		{object}	httpsuite.APIError	"Bad Request"
// @Failure		500		{object}	httpsuite.APIError	"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues/{queue-id} [get]
func (rc *APIService) GetQueuesByNameHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	queue := chi.URLParam(r, "queue")
	if queue == "" {
		httpsuite.WriteJSONError(w, "missing queue name", http.StatusBadRequest)
		return
	}

	queues, err := rc.RMQClient.GetQueueByName(vhost, queue)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queue details", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queue details", http.StatusOK, &queues)
}
