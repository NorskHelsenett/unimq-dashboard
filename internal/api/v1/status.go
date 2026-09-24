package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
)

// @Summary		Get checker status
// @Description	Get the current status of the notification checker
// @Tags			Checker
// @Produce		json
// @Success		200	{object}	notify.CheckerStatus
// @Failure		400	{object}	httpsuite.ErrorResponse
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		503	{object}	httpsuite.ErrorResponse
// @Router			/v1/status [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetCheckerStatusHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

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
