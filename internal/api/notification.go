package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get Notifications
// @Description	Get all notification vhosts and settings
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	[]models.VhostNotification
// @Failure		502	{object}	httpsuite.APIError
// @Router			/v1/notifications [get]
func (rc *APIService) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	notifications, err := rc.DB.GetNotificationsAll(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification configs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notifications", http.StatusOK, &notifications)
}

// @Summary		Get Notification on Vhost
// @Description	Get notification rules and settings for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.VhostNotification
// @Failure		400			{object}	httpsuite.APIError
// @Failure		502			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name} [get]
func (rc *APIService) GetNotificationsVhostHandler(w http.ResponseWriter, r *http.Request) {

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
// @Description	Deletes the notification configuration for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	string
// @Failure		400			{object}	httpsuite.APIError
// @Failure		502			{object}	httpsuite.APIError
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
