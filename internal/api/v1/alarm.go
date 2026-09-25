package api

import (
	"errors"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/requesthelper"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary		Get Alarm History
// @Description	Get alarm history for all rules
// @Tags			Alarms
// @Produce		json
// @Success		200	{array}		[]models.AlarmEntry
// @Failure		400	{object}	httpsuite.ErrorResponse
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		502	{object}	httpsuite.ErrorResponse
// @Router			/v1/alarms [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetAlarmHistoryAllHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	alarms, err := rc.DB.GetAlarmsAll(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("error fetching alarm history"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notification history on vhost", http.StatusOK, &alarms)
}

// @Summary		Get Alarm History for a rule
// @Description	Get alarm history for a specific rule
// @Tags			Alarms
// @Produce		json
// @Param			rule-id	path		string	true	"Rule ID"
// @Success		200		{array}		[]models.AlarmEntry
// @Failure		400		{object}	httpsuite.ErrorResponse
// @Failure		401		{object}	httpsuite.ErrorResponse
// @Failure		403		{object}	httpsuite.ErrorResponse
// @Failure		404		{object}	httpsuite.ErrorResponse
// @Failure		502		{object}	httpsuite.ErrorResponse
// @Router			/v1/alarms/{rule-id} [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetAlarmHistoryHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	ruleID, err := requesthelper.ReadRuleIDFromRequest(r)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}

	alarms, err := rc.DB.GetAlarm(r.Context(), ruleID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithErrorMessage("no alarm history found for rule ID: "+ruleID),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch alarm history for rule ID: "+ruleID),
		)
		return
	}

	if alarms == nil {
		httpsuite.WriteJSONError(w,
			http.StatusNotFound,
			httpsuite.WithErrorMessage("no alarm history found for rule ID: "+ruleID),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notification history on rule ID "+ruleID, http.StatusOK, alarms)
}
