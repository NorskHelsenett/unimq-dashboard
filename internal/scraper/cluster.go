package scraper

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func (rc *RestClient) GetClusterHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := rc.RMQClient.GetClusterStats()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching cluster stats", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered cluster stats", http.StatusOK, stats)
}
