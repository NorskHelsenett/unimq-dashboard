package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Add a new notification rule
// @Description	Add a new notification rule for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost	path		string				true	"Vhost Name"
// @Param			rule	body		models.AlarmRule	true	"Notification Rule Object"
// @Success		303		{string}	string				"Redirect to notifications page"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules [post]
func (rc *APIService) AddNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	var rule models.AlarmRule
	err := httpsuite.ReadResponse(r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = rc.DB.AddNotificationRule(r.Context(), vhost, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

// @Summary		Delete a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost		path		string	true	"Vhost Name"
// @Param			recipient	path		string	true	"Notification Rule ID"
// @Success		303			{string}	string	"Redirect to notifications page"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [delete]
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

// @Summary		Delete a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost		path		string					true	"Vhost Name"
// @Param			recipient	path		string					true	"Notification Rule ID"
// @Param			rule		body		models.AlarmRuleUpdate	true	"Updated Notification Rule Object"
// @Success		200			{string}	string					"Rule updated successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [put]
func (rc *APIService) UpdateNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	var rule models.AlarmRuleUpdate
	err := httpsuite.ReadResponse(r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = rc.DB.UpdateNotificationRuleThreshold(r.Context(), vhost, id, rule.Threshold)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule threshold: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.DB.UpdateNotificationRuleMessage(r.Context(), vhost, id, rule.Message)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary		Toggle a notification rule
// @Description	Enable or disable a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost	path		string	true	"Vhost Name"
// @Param			rule	path		string	true	"Notification Rule ID"
// @Success		303		{string}	string	"Redirect to notifications page"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/toggle [post]
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
