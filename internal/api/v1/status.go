package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get checker status
// @Description	Get the current status of the notification checker
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	notify.CheckerStatus
// @Failure		503	{object}	httpsuite.ErrorResponse
// @Router			/v1/checker/status [get]
// @security		bearer
func (rc *APIService) GetCheckerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if rc.Checker == nil {
		httpsuite.WriteJSONError(w,
			http.StatusServiceUnavailable,
			httpsuite.WithErrorMessage("checker is not available"),
		)
		return
	}
	status := rc.Checker.GetStatus()
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &status)
}
