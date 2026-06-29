package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Notifications
// @Description	Get notification rules and settings for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200		{object}	models.VhostNotification
// @Failure		400		{object}	httpsuite.APIError
// @Failure		502		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name} [get]
func (rc *APIService) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	notification, err := rc.DB.GetNotification(r.Context(), vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notifications on vhost", http.StatusOK, notification)
}

// @Summary		Delete Notifications
// @Description	Delete all notification rules and settings for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200		{object}	httpsuite.EmptyResponse
// @Failure		400		{object}	httpsuite.APIError
// @Failure		502		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name} [delete]
func (rc *APIService) DeleteNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteNotification(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting notification config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Deleted notifications on vhost", http.StatusOK)
}

// @Summary		Update Notification Rule Message
// @Description	Update the custom message template for a specific notification rule
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Param			rule-id	path		string	true	"Notification Rule ID"
// @Param			message	formData	string	true	"New Message Template"
// @Success		303		{string}	string	"Redirect to notification rule page with success message"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/message [post]
func (rc *APIService) UpdateNotificationsMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	err := rc.DB.UpdateNotificationRuleMessage(r.Context(), vhost, id, r.FormValue("message"))
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "notification rule message updated successfully", http.StatusOK)
}

// @Summary		Test Notification Rule
// @Description	Send a test notification using the specified rule to verify its configuration
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string							true	"Vhost Name"
// @Param			rule-id	path		string							true	"Notification Rule ID"
// @Success		200		{object}	models.TestNotificationResponse	"Test notification sent successfully"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/test [post]
func (rc *APIService) TestNotificationsHandler(w http.ResponseWriter, r *http.Request) {
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

	vhostobject, err := rc.DB.GetVhost(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}

	subject := "[UniMQ TEST] " + rule.Name + " — " + vhost
	body := "This is a test message from UniMQ.\n\n" + rule.BuildMessage(vhost)
	if err := notificationhelper.SendWebhooks(vhostobject.WebhookURLs(), subject, body); err != nil {
		httpsuite.WriteJSONError(w, "error sending test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.TestNotificationResponse{
		Success: true,
		Message: "Test notification sent!",
	}

	httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
}
