package scraper

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, _ := rc.GetVhosts()
	selected := r.URL.Query().Get("vhost")
	if selected == "" && len(vhosts) > 0 {
		selected = vhosts[0]
	}
	vc := templating.NotifyStore.GetVhostCopy(selected)
	data := models.NotifPageData{
		Vhosts:     vhosts,
		Selected:   selected,
		Recipients: vc.Recipients,
		Rules:      vc.Rules,
	}
	if err := templating.NotifTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (rc *RestClient) PostNotificationsAddRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	templating.NotifyStore.AddRecipient(vhost, r.FormValue("name"), r.FormValue("url"), r.FormValue("type"))
	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsDeleteRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	templating.NotifyStore.DeleteRecipient(vhost, r.FormValue("id"))
	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsAddRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	threshold, _ := strconv.ParseFloat(r.FormValue("threshold"), 64)
	rule := models.AlarmRule{
		Name:      r.FormValue("name"),
		Type:      r.FormValue("type"),
		QueueName: r.FormValue("queue_name"),
		Threshold: threshold,
		Message:   r.FormValue("message"),
	}
	templating.NotifyStore.AddRule(vhost, rule)
	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsDeleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	templating.NotifyStore.DeleteRule(vhost, r.FormValue("id"))
	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) PostNotificationsUpdateRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	id := r.FormValue("id")
	threshold, _ := strconv.ParseFloat(r.FormValue("threshold"), 64)
	templating.NotifyStore.UpdateRule(vhost, id, r.FormValue("message"), threshold)
	w.WriteHeader(http.StatusOK)
}

func (rc *RestClient) PostNotificationsToggleRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.FormValue("vhost")
	templating.NotifyStore.ToggleRule(vhost, r.FormValue("id"))
	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

func (rc *RestClient) NotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	id := r.URL.Query().Get("id")
	rule, ok := templating.NotifyStore.GetRuleCopy(vhost, id)
	if !ok {
		http.Error(w, "Regel ikke funnet", http.StatusNotFound)
		return
	}
	data := struct {
		Vhost string
		Rule  models.AlarmRule
		Msg   string
	}{vhost, rule, r.URL.Query().Get("msg")}
	if err := templating.NotifRuleTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (rc *RestClient) NotificationsUpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		id := r.FormValue("id")
		templating.NotifyStore.UpdateMessage(vhost, id, r.FormValue("message"))
		http.Redirect(w, r, "/notifications/rule?vhost="+url.QueryEscape(vhost)+"&id="+id+"&msg=saved", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func (rc *RestClient) NotificationsTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		id := r.FormValue("id")
		rule, ok := templating.NotifyStore.GetRuleCopy(vhost, id)
		vc := templating.NotifyStore.GetVhostCopy(vhost)
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Rule not found"})
			return
		}
		if len(vc.WebhookURLs()) == 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "no_recipients", "message": "No recipients configured for this vhost"})
			return
		}
		subject := "[UniMQ TEST] " + rule.Name + " — " + vhost
		body := "Dette er en test-varsling fra UniMQ.\n\n" + rule.BuildMessage(vhost)
		if err := templating.NotifyStore.SendWebhooks(vc.WebhookURLs(), subject, body); err != nil {
			log.Printf("notify test webhook failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Failed to send webhook: " + err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "sent", "message": "Test notification sent!"})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (rc *RestClient) NotificationsLogsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	entries := templating.LogStore.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
