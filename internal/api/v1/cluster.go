package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Cluster Stats
// @Description	Get overall cluster statistics and health information
// @Tags			Cluster
// @Produce		json
// @Success		200	{object}	models.ClusterStats
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/cluster [get]
// @security		bearer
func (rc *APIService) GetClusterHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	stats, err := rc.RMQClient.GetClusterStats()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch cluster stats"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered cluster stats", http.StatusOK, stats)
}
