package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func (rc *APIService) VhostsHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, err := rc.RMQClient.GetVhosts()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhosts", http.StatusBadGateway)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhosts)
}

type vhostResponse struct {
	Vhost   *models.Vhost        `json:"vhost"`
	Metrics *models.VhostMetrics `json:"metrics"`
}

func (rc *APIService) VhostHandler(w http.ResponseWriter, r *http.Request) {
	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w, "vhost name is required", http.StatusBadRequest)
		return
	}

	vhostData, err := rc.RMQClient.GetVhost(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost data", http.StatusBadGateway)
		return
	}

	metrics, err := rc.RMQClient.GetMetrics(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost metrics", http.StatusBadGateway)
		return
	}

	response := vhostResponse{
		Vhost:   vhostData,
		Metrics: metrics,
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &response)
}
