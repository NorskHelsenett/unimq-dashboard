package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary		Get Alarm History
// @Description	Get alarm history for all rules
// @Tags			Alarms
// @Produce		json
// @Success		200	{array}		[]models.AlarmEntry
// @Failure		400	{object}	httpsuite.APIError
// @Failure		502	{object}	httpsuite.APIError
// @Router			/v1/alarms [get]
// @security		bearer
func (rc *APIService) GetAlarmHistoryAllHandler(w http.ResponseWriter, r *http.Request) {

	alarms, err := rc.DB.GetAlarmsAll(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching alarm history: "+err.Error(), http.StatusInternalServerError)
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
// @Failure		400		{object}	httpsuite.APIError
// @Failure		404		{object}	httpsuite.APIError
// @Failure		502		{object}	httpsuite.APIError
// @Router			/v1/alarms/{rule-id} [get]
// @security		bearer
func (rc *APIService) GetAlarmHistoryHandler(w http.ResponseWriter, r *http.Request) {

	ruleID := chi.URLParam(r, "rule-id")
	if ruleID == "" {
		httpsuite.WriteJSONError(w, "missing required rule-id parameter", http.StatusBadRequest)
		return
	}

	eRuleID, err := url.QueryUnescape(ruleID)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding rule-id  name", http.StatusBadRequest)
		return
	}

	alarms, err := rc.DB.GetAlarm(r.Context(), eRuleID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpsuite.WriteJSONError(w, "no alarm history found for rule ID: "+eRuleID, http.StatusNotFound)
			return
		}
		httpsuite.WriteJSONError(w, "error fetching alarm history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if alarms == nil {
		httpsuite.WriteJSONError(w, "no alarm history found for rule ID: "+eRuleID, http.StatusNotFound)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notification history on rule ID "+eRuleID, http.StatusOK, alarms)
}
