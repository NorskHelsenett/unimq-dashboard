package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Add a new notification rule
// @Description	Add a new notification rule for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			rule		body		models.PostAlarmRule	true	"Notification Rule Object"
// @Success		201			{object}	string					"Rule added successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		404			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules [post]
func (rc *APIService) AddNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	var rule models.PostAlarmRule
	err := httpsuite.ReadResponse(r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = rc.ensureNotificationHostExists(r.Context(), vhost)
	if err != nil {
		if errors.Is(err, rabbitmq.ErrVhostNotFound) {
			httpsuite.WriteJSONError(w, "vhost not found: "+err.Error(), http.StatusNotFound)
			return
		}
		httpsuite.WriteJSONError(w, "error validating vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := rule.ToAlarmRule()
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid rule data: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = rc.DB.AddNotificationRule(r.Context(), vhost, out)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Rule added successfully", http.StatusCreated)
}

// @Summary		Delete a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Param			rule-id		path		string	true	"Notification Rule ID"
// @Success		200			{string}	string	"Rule deleted successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [delete]
func (rc *APIService) DeleteNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required rule id parameter", http.StatusBadRequest)
		return
	}
	err := rc.DB.DeleteNotificationRule(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Rule deleted successfully", http.StatusOK)
}

// @Summary		Update a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			rule-id		path		string					true	"Notification Rule ID"
// @Param			rule		body		models.AlarmRuleUpdate	true	"Updated Notification Rule Object"
// @Success		200			{string}	string					"Rule updated successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [Post]
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

	httpsuite.SendResponse(r.Context(), w, "Rule updated successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Toggle a notification rule
// @Description	Enable or disable a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Param			rule-id		path		string	true	"Notification Rule ID"
// @Success		200			{string}	string	"Rule toggled successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
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

	httpsuite.SendResponse(r.Context(), w, "Rule toggled successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Test Notification Rule
// @Description	Send a test notification using the specified rule to verify its configuration
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string							true	"Vhost Name"
// @Param			rule-id		path		string							true	"Notification Rule ID"
// @Success		200			{object}	models.TestNotificationResponse	"Test notification sent successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/test [post]
func (rc *APIService) TestNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, database.ErrVhostNotFound) {
			httpsuite.WriteJSONError(w, fmt.Sprintf("vhost %v not found", vhost), http.StatusNotFound)
			return
		}
		if errors.Is(err, database.ErrNotificationRuleNotFound) {
			httpsuite.WriteJSONError(w, fmt.Sprintf("notification rule not found with id %v on vhost %v", id, vhost), http.StatusNotFound)
			return
		}
		httpsuite.WriteJSONError(w, "error fetching rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	vhostobject, err := rc.DB.GetVhost(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}

	subject := "[UniMQ TEST] " + rule.Name + " — " + vhost
	body := "This is a test message from UniMQ.\n\n" + rule.BuildMessage(vhost)

	urls := vhostobject.WebhookURLs()
	if len(urls) == 0 {
		httpsuite.WriteJSONError(w, "no webhook URLs configured for vhost "+vhost, http.StatusBadRequest)
		return
	}

	err = notificationhelper.SendWebhooks(urls, subject, body)
	if err != nil {
		httpsuite.WriteJSONError(w, "error sending test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.TestNotificationResponse{
		Success: true,
		Message: "Test notification sent!",
	}

	httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
}
