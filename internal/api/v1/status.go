package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func (rc *APIService) GetCheckerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if rc.Checker == nil {
		httpsuite.WriteJSONError(w, "checker not available", http.StatusServiceUnavailable)
		return
	}
	status := rc.Checker.GetStatus()
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &status)
}
