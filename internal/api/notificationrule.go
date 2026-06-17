package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

func (rc *APIService) AddNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	threshold, err := strconv.ParseFloat(r.FormValue("threshold"), 64)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid threshold value: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		httpsuite.WriteJSONError(w, "missing required name parameter", http.StatusBadRequest)
		return
	}

	typ := r.FormValue("type")
	if typ == "" {
		httpsuite.WriteJSONError(w, "missing required type parameter", http.StatusBadRequest)
		return
	}

	queueName := r.FormValue("queue_name")
	if typ == "queue_length" && queueName == "" {
		httpsuite.WriteJSONError(w, "missing required queue_name parameter for queue_length type", http.StatusBadRequest)
		return
	}
	message := r.FormValue("message")
	if message == "" {
		httpsuite.WriteJSONError(w, "missing required message parameter", http.StatusBadRequest)
		return
	}

	rule := models.AlarmRule{
		Name:      name,
		Type:      models.AlarmType(typ),
		QueueName: queueName,
		Threshold: threshold,
		Message:   message,
	}

	err = rc.DB.AddNotificationRule(r.Context(), vhost, rule)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *APIService) DeleteNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required recipient id parameter", http.StatusBadRequest)
		return
	}
	err := rc.DB.DeleteNotificationRule(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *APIService) PostNotificationsUpdateRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	thresholdStr := r.FormValue("threshold")
	if thresholdStr == "" {
		httpsuite.WriteJSONError(w, "missing required threshold parameter", http.StatusBadRequest)
		return
	}

	threshold, err := strconv.ParseFloat(r.FormValue("threshold"), 64)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid threshold value: "+err.Error(), http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		httpsuite.WriteJSONError(w, "missing required message parameter", http.StatusBadRequest)
		return
	}

	err = rc.DB.UpdateNotificationRuleThreshold(r.Context(), vhost, id, threshold)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule threshold: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.DB.UpdateNotificationRuleMessage(r.Context(), vhost, id, message)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (rc *APIService) ToggleNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required rule parameter", http.StatusBadRequest)
		return
	}

	rule, err := rc.DB.GetNotificationRule(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if rule == nil {
		httpsuite.WriteJSONError(w, "rule not found", http.StatusNotFound)
		return
	}

	err = rc.DB.ToggleNotificationRule(r.Context(), vhost, id, !rule.Enabled)
	if err != nil {
		httpsuite.WriteJSONError(w, "error toggling rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *APIService) NotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	rule, err := rc.DB.GetNotificationRule(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Vhost string
		Rule  models.AlarmRule
	}{vhost, *rule}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &data)
}
