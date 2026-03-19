package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/maintenance"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

var (
	indexTmpl      = template.Must(template.New("index.html").Funcs(template.FuncMap{"div": func(a, b int) int { return a / b }}).ParseFiles("web/templates/index.html"))
	queueTmpl      = template.Must(template.ParseFiles("web/templates/queue.html"))
	maintTmpl      = template.Must(template.ParseFiles("web/templates/maintenance.html"))
	maintAdminTmpl = template.Must(template.ParseFiles("web/templates/maintenance_admin.html"))
	notifTmpl      = template.Must(template.ParseFiles("web/templates/notifications.html"))
	notifRuleTmpl  = template.Must(template.ParseFiles("web/templates/notification_rule.html"))

	maintStore  *maintenance.Store
	notifyStore *notify.Store
)

// ── page data structs ────────────────────────────────────────────────────────

type pageData struct {
	Vhosts   []string
	Selected string
	Metrics  *scraper.VhostMetrics
	Limits   scraper.Limits
}

type rangeOption struct {
	Label string
	Value string
}

var timeRanges = []rangeOption{
	{"5m", "5m"}, {"1h", "1h"}, {"6h", "6h"}, {"24h", "24h"}, {"7d", "7d"},
}

var rangeDurations = map[string]time.Duration{
	"5m": 5 * time.Minute, "1h": time.Hour,
	"6h": 6 * time.Hour, "24h": 24 * time.Hour, "7d": 7 * 24 * time.Hour,
}

// ── overview ─────────────────────────────────────────────────────────────────

func indexHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, err := scraper.GetVhosts()
	if err != nil {
		http.Error(w, "Could not reach RabbitMQ: "+err.Error(), http.StatusBadGateway)
		return
	}
	selected := r.URL.Query().Get("vhost")
	if selected == "" && len(vhosts) > 0 {
		selected = vhosts[0]
	}
	data := pageData{Vhosts: vhosts, Selected: selected, Limits: scraper.DefaultLimits}
	if selected != "" {
		if m, err := scraper.GetMetrics(selected); err == nil {
			data.Metrics = m
		}
	}
	if err := indexTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

// ── queue detail ─────────────────────────────────────────────────────────────

func queueHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	queue := r.URL.Query().Get("name")
	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "1h"
	}
	since, ok := rangeDurations[rangeStr]
	if !ok {
		since = time.Hour
		rangeStr = "1h"
	}
	samples, err := prom.QueryRange(prom.RangeOptions{Vhost: vhost, Queue: queue, Since: since, Step: prom.StepFor(since)})
	if err != nil {
		log.Printf("prometheus error: %v", err)
	}
	samplesJSON := "[]"
	if len(samples) > 0 {
		b, _ := json.Marshal(samples)
		samplesJSON = string(b)
	}
	data := struct {
		Vhost, Queue, SelectedRange string
		Ranges                      []rangeOption
		SamplesJSON                 template.JS
		NoData                      bool
	}{vhost, queue, rangeStr, timeRanges, template.JS(samplesJSON), len(samples) == 0}
	if err := queueTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

// ── APIs ─────────────────────────────────────────────────────────────────────

func queuesAPIHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		http.Error(w, "missing vhost", http.StatusBadRequest)
		return
	}
	details, err := scraper.GetQueueDetails(vhost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

func clusterAPIHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := scraper.GetClusterStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ── maintenance ───────────────────────────────────────────────────────────────

func maintenanceHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Scheduled []maintenance.Entry
		History   []maintenance.Entry
	}{maintStore.Scheduled(), maintStore.History()}
	if err := maintTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func maintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {
	data := struct{ Entries []maintenance.Entry }{maintStore.All()}
	if err := maintAdminTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func maintenanceAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		maintStore.Add(r.FormValue("description"), r.FormValue("start"), r.FormValue("end"))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func maintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		maintStore.SetStatus(r.FormValue("id"), r.FormValue("status"))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func maintenanceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		maintStore.Delete(r.FormValue("id"))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

// ── notifications ─────────────────────────────────────────────────────────────

type notifPageData struct {
	Vhosts     []string
	Selected   string
	Recipients []notify.Recipient
	Rules      []notify.AlarmRule
}

func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, _ := scraper.GetVhosts()
	selected := r.URL.Query().Get("vhost")
	if selected == "" && len(vhosts) > 0 {
		selected = vhosts[0]
	}
	vc := notifyStore.GetVhostCopy(selected)
	data := notifPageData{
		Vhosts:     vhosts,
		Selected:   selected,
		Recipients: vc.Recipients,
		Rules:      vc.Rules,
	}
	if err := notifTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func notificationsAddRecipientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		notifyStore.AddRecipient(vhost, r.FormValue("name"), r.FormValue("url"), r.FormValue("type"))
		http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func notificationsDeleteRecipientHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		notifyStore.DeleteRecipient(vhost, r.FormValue("id"))
		http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}


func notificationsAddRuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		threshold, _ := strconv.ParseFloat(r.FormValue("threshold"), 64)
		rule := notify.AlarmRule{
			Name:      r.FormValue("name"),
			Type:      r.FormValue("type"),
			QueueName: r.FormValue("queue_name"),
			Threshold: threshold,
			Message:   r.FormValue("message"),
		}
		notifyStore.AddRule(vhost, rule)
		http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func notificationsDeleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		notifyStore.DeleteRule(vhost, r.FormValue("id"))
		http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func notificationsToggleRuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		notifyStore.ToggleRule(vhost, r.FormValue("id"))
		http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func notificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	id := r.URL.Query().Get("id")
	rule, ok := notifyStore.GetRuleCopy(vhost, id)
	if !ok {
		http.Error(w, "Regel ikke funnet", http.StatusNotFound)
		return
	}
	data := struct {
		Vhost string
		Rule  notify.AlarmRule
		Msg   string
	}{vhost, rule, r.URL.Query().Get("msg")}
	if err := notifRuleTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func notificationsUpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		id := r.FormValue("id")
		notifyStore.UpdateMessage(vhost, id, r.FormValue("message"))
		http.Redirect(w, r, "/notifications/rule?vhost="+url.QueryEscape(vhost)+"&id="+id+"&msg=saved", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func notificationsTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		vhost := r.FormValue("vhost")
		id := r.FormValue("id")
		rule, ok := notifyStore.GetRuleCopy(vhost, id)
		vc := notifyStore.GetVhostCopy(vhost)
		redirect := "/notifications/rule?vhost=" + url.QueryEscape(vhost) + "&id=" + id
		if ok && len(vc.WebhookURLs()) > 0 {
			subject := "[UniMQ TEST] " + rule.Name + " — " + vhost
			body := "Dette er en test-varsling fra UniMQ.\n\n" + notify.BuildMessage(rule, vhost)
			if err := notifyStore.SendWebhooks(vc.WebhookURLs(), subject, body); err != nil {
				log.Printf("notify test webhook failed: %v", err)
				http.Redirect(w, r, redirect+"&msg=error", http.StatusSeeOther)
				return
			}
		}
		http.Redirect(w, r, redirect+"&msg=sent", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	var err error

	maintStore, err = maintenance.NewStore("data/maintenance.json")
	if err != nil {
		log.Fatalf("could not load maintenance store: %v", err)
	}

	notifyStore, err = notify.NewStore("data/notifications.json")
	if err != nil {
		log.Fatalf("could not load notification store: %v", err)
	}

	// Start background alarm checker — runs every 60 seconds.
	notify.StartChecker(notifyStore, maintStore, 60*time.Second)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/queue", queueHandler)
	http.HandleFunc("/maintenance", maintenanceHandler)
	http.HandleFunc("/maintenance/admin", maintenanceAdminHandler)
	http.HandleFunc("/maintenance/add", maintenanceAddHandler)
	http.HandleFunc("/maintenance/status", maintenanceStatusHandler)
	http.HandleFunc("/maintenance/delete", maintenanceDeleteHandler)
	http.HandleFunc("/notifications", notificationsHandler)
	http.HandleFunc("/notifications/recipients/add", notificationsAddRecipientHandler)
	http.HandleFunc("/notifications/recipients/delete", notificationsDeleteRecipientHandler)
	http.HandleFunc("/notifications/rules/add", notificationsAddRuleHandler)
	http.HandleFunc("/notifications/rules/delete", notificationsDeleteRuleHandler)
	http.HandleFunc("/notifications/rules/toggle", notificationsToggleRuleHandler)
	http.HandleFunc("/notifications/rules/message", notificationsUpdateMessageHandler)
	http.HandleFunc("/notifications/rules/test", notificationsTestHandler)
	http.HandleFunc("/notifications/rule", notificationsRuleHandler)
	http.HandleFunc("/api/queues", queuesAPIHandler)
	http.HandleFunc("/api/cluster", clusterAPIHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// CONFIG: Bytt port hvis 8080 er opptatt eller du ønsker en annen port.
	//         Husk å oppdatere URL-en i eventuelle reverse proxy-oppsett (nginx, etc.).
	log.Println("Dashboard running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
