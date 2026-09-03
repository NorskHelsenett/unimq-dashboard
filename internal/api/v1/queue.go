package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Queues for a specific vhost
// @Description	Fetches a list of all queues in a specified virtual host.
// @Tags			Queues
// @Produce		json
// @Param			vhost-name	path		string						true	"Virtual Host"
// @Success		200			{object}	[]models.QueueAPIResponse	"HTML page with queue metrics"
// @Failure		400			{object}	httpsuite.APIError			"Bad Request"
// @Failure		404			{object}	httpsuite.APIError			"Not Found"
// @Failure		500			{object}	httpsuite.APIError			"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues [get]
func (rc *APIService) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	queues, err := rc.RMQClient.GetQueue(eVhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queues. "+err.Error(), http.StatusNotFound)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queues", http.StatusOK, &queues)
}

// @Summary		Get a specific queue to a specified virtual host
// @Description	Fetches details of all queues in a specified virtual host.
// @Tags			Queues
// @Produce		json
// @Param			vhost-name	path		string				true	"Virtual Host"
// @Param			queue-id	path		string				true	"Queue Name"
// @Success		200			{array}		models.QueueDetail	"List of queue details"
// @Failure		400			{object}	httpsuite.APIError	"Bad Request"
// @Failure		404			{object}	httpsuite.APIError	"Not Found"
// @Failure		500			{object}	httpsuite.APIError	"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues/{queue-id} [get]
func (rc *APIService) GetQueuesByNameHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	queue := chi.URLParam(r, "queue")
	if queue == "" {
		httpsuite.WriteJSONError(w, "missing queue name", http.StatusBadRequest)
		return
	}

	eQueue, err := url.QueryUnescape(queue)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding queue name", http.StatusBadRequest)
		return
	}

	queues, err := rc.RMQClient.GetQueueByName(eVhost, eQueue)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queue details. "+err.Error(), http.StatusNotFound)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queue details", http.StatusOK, &queues)
}
