package scraper

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

// TODO: Use the Validator package to validate input requirements.

func (rc *RestClient) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		// slog.ErrorContext(r.Context(), "missing vhost parameter")
		return
	}
	vhosts, err := rc.GetVhosts()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusBadGateway)
		// slog.ErrorContext(r.Context(), "error fetching vhost", "vhost", vhost, "error", err)
		return
	}

	vhostobject, err := rc.DB.GetNotification(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification config: "+err.Error(), http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error fetching notification config", "vhost", vhost, "error", err)
		return
	}
	data := models.NotifPageData{
		Vhosts:     vhosts,
		Selected:   vhost,
		Recipients: vhostobject.Recipients,
		Rules:      vhostobject.Rules,
	}
	if err := templating.NotifTmpl.Execute(w, data); err != nil {
		httpsuite.WriteJSONError(w, "error rendering template: "+err.Error(), http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error rendering template", "vhost", vhost, "error", err)
	}
}

func (rc *RestClient) PostNotificationsAddRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		httpsuite.WriteJSONError(w, "missing required name parameter", http.StatusBadRequest)
		return
	}

	urlP := r.FormValue("url")
	if urlP == "" {
		httpsuite.WriteJSONError(w, "missing required url parameter", http.StatusBadRequest)
		return
	}

	typ := r.FormValue("type")
	if typ == "" {
		httpsuite.WriteJSONError(w, "missing required type parameter", http.StatusBadRequest)
		return
	}

	if slices.Contains(models.GetRuleTypes(), models.RuleType(typ)) == false {
		httpsuite.WriteJSONError(w, "invalid type parameter, expected one of "+models.GetRuleTypesString(), http.StatusBadRequest)
		return
	}

	id := strconv.FormatInt(int64(len(templating.NotifyStore.GetVhostCopy(vhost).Recipients)+1), 10)

	recipient := models.Recipient{
		ID:   id,
		Name: name,
		URL:  urlP,
		Type: models.ParseRuleType(typ),
	}

	err := rc.DB.AddNotificationRecipient(r.Context(), vhost, recipient)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsDeleteRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteNotificationRecipient(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsAddRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
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
		Type:      typ,
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

func (rc *RestClient) PostNotificationsDeleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}
	err := rc.DB.DeleteNotificationRule(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting rule: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsUpdateRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
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

func (rc *RestClient) PostNotificationsToggleRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
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

func (rc *RestClient) NotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
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
	if err := templating.NotifRuleTmpl.Execute(w, data); err != nil {
		httpsuite.WriteJSONError(w, "error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (rc *RestClient) PostNotificationsUpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	err := rc.DB.UpdateNotificationRuleMessage(r.Context(), vhost, id, r.FormValue("message"))
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating rule message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications/rule?vhost="+url.QueryEscape(vhost)+"&id="+id+"&msg=saved", http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsTestHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
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
	if err := templating.NotifyStore.SendWebhooks(vhostobject.WebhookURLs(), subject, body); err != nil {
		httpsuite.WriteJSONError(w, "error sending test notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "sent",
		"message": "Test notification sent!",
	}

	httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
}

func (rc *RestClient) NotificationsLogsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	entries, err := rc.DB.GetNotification(r.Context(), id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching notification logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Fetched logs for rule "+id, http.StatusOK, &entries)

}
