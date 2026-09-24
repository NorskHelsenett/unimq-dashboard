package rmq

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
)

// @Summary		Get Queues for a specific vhost
// @Description	Fetches a list of all queues in a specified virtual host.
// @Tags			Queues
// @Produce		json
// @Param			vhost-name	path		string					true	"Virtual Host"
// @Success		200			{object}	[]models.RMQQueue		"HTML page with queue metrics"
// @Failure		400			{object}	httpsuite.ErrorResponse	"Bad Request"
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		404			{object}	httpsuite.ErrorResponse	"Not Found"
// @Failure		500			{object}	httpsuite.ErrorResponse	"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing vhost parameter"),
		)
		return
	}

	queues, err := rc.RMQClient.GetQueue(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusNotFound,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch queues for vhost"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queues", http.StatusOK, &queues)
}

// @Summary		Get a specific queue to a specified virtual host
// @Description	Fetches details of all queues in a specified virtual host.
// @Tags			Queues
// @Produce		json
// @Param			vhost-name	path		string					true	"Virtual Host"
// @Param			queue-id	path		string					true	"Queue Name"
// @Success		200			{array}		models.QueueDetail		"List of queue details"
// @Failure		400			{object}	httpsuite.ErrorResponse	"Bad Request"
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		404			{object}	httpsuite.ErrorResponse	"Not Found"
// @Failure		500			{object}	httpsuite.ErrorResponse	"Internal Server Error"
// @Router			/v1/vhosts/{vhost-name}/queues/{queue-id} [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) GetQueuesByNameHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing vhost parameter"),
		)
		return
	}

	queue := chi.URLParam(r, "queue-id")
	if queue == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing queue name"),
		)
		return
	}

	eQueue, err := url.QueryUnescape(queue)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to decode queue name"),
			httpsuite.WithInternalErrorMessage("failed to decode queue name: "+queue),
		)
		return
	}

	queues, err := rc.RMQClient.GetQueueByName(vhost, eQueue)
	if err != nil {
		if errors.Is(err, rabbitmq.ErrQueueNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("queue not found for vhost"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch queue details for vhost"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queue details", http.StatusOK, &queues)
}
