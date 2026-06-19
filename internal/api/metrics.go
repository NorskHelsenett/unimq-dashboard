package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @ Summary Get Vhost Metrics
// @ Description Get real-time metrics for a specific vhost, including queue lengths, message rates, and resource usage
// @ Tags Metrics
// @ Produce json
// @ Param vhost path string true "Vhost Name"
// @ Success 200 {object} models.VhostMetrics
// @ Failure 400 {object} httpsuite.APIError
// @ Failure 502 {object} httpsuite.APIError
// @ Router /v1/vhosts/{vhost}/metrics [get]
func (rc *APIService) MetricHandler(w http.ResponseWriter, r *http.Request) {
	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w, "vhost name is required", http.StatusBadRequest)
		return
	}
	metrics, err := rc.RMQClient.GetMetrics(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching metrics for vhost", http.StatusBadGateway)
		return
	}
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &metrics)
}
