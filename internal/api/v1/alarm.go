package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Alarm History
// @Description	Get alarm history for all vhosts
// @Tags			Alarms
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{array}		[]models.AlarmEntry
// @Failure		400			{object}	httpsuite.APIError
// @Failure		502			{object}	httpsuite.APIError
// @Router			/v1/alarms [get]
func (rc *APIService) GetAlarmHistoryAllHandler(w http.ResponseWriter, r *http.Request) {

	alarms, err := rc.DB.GetAlarmsAll(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching alarm history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notification history on vhost", http.StatusOK, &alarms)
}

// @Summary		Get Alarm History for Vhost
// @Description	Get alarm history for a specific vhost
// @Tags			Alarms
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{array}		[]models.AlarmEntry
// @Failure		400			{object}	httpsuite.APIError
// @Failure		502			{object}	httpsuite.APIError
// @Router			/v1/alarms/{vhost-name} [get]
func (rc *APIService) GetAlarmHistoryHandler(w http.ResponseWriter, r *http.Request) {

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	alarms, err := rc.DB.GetAlarm(r.Context(), eVhost)

	httpsuite.SendResponse(r.Context(), w, "Gathered notification history on vhost", http.StatusOK, alarms)
}
