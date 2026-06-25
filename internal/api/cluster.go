package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Cluster Stats
// @Description	Get overall cluster statistics and health information
// @Tags		Cluster
// @Produce		json
// @Success		200	{object}	models.ClusterStats
// @Failure		500	{object}	httpsuite.APIError
// @Router		/v1/cluster [get]
func (rc *APIService) GetClusterHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := rc.RMQClient.GetClusterStats()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching cluster stats", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered cluster stats", http.StatusOK, stats)
}
