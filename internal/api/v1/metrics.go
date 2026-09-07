package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Vhost Metrics
// @Description	Get real-time metrics for a specific vhost, including queue lengths, message rates, and resource usage
// @Tags			Vhosts
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.VhostMetrics
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts/{vhost-name}/metrics [get]
// @security		bearer
func (rc *APIService) MetricHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("vhost name is required"),
		)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to decode vhost name"),
			httpsuite.WithInternalErrorMessage("error decoding vhost name: "+vhost),
		)
		return
	}

	metrics, err := rc.RMQClient.GetMetrics(eVhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch metrics for vhost"),
		)
		return
	}
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &metrics)
}
