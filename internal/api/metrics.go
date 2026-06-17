package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

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
