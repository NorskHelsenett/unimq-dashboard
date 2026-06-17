package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

type notificationResponse struct {
	Vhost         *models.Vhost               `json:"vhost"`
	Notifications *database.VhostNotification `json:"notifications"`
}

func (rc *APIService) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	vhost, err := rc.RMQClient.GetVhost(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusBadGateway)
		return
	}

	vhostobject, err := rc.DB.GetNotification(r.Context(), vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := notificationResponse{
		Vhost:         vhost,
		Notifications: vhostobject,
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &data)
}

func (rc *APIService) PostNotificationsUpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/notifications/rule?vhost="+url.QueryEscape(vhost)+"&id="+id+"&msg=saved", http.StatusSeeOther)
}

func (rc *APIService) PostNotificationsTestHandler(w http.ResponseWriter, r *http.Request) {
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

	response := map[string]string{
		"status":  "sent",
		"message": "Test notification sent!",
	}

	httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
}

func (rc *APIService) NotificationsLogsHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "vhost")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	entries, err := rc.DB.GetNotification(r.Context(), id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Fetched logs for rule "+id, http.StatusOK, &entries)

}
