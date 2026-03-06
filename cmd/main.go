package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

var (
	indexTmpl = template.Must(
		template.New("index.html").
			Funcs(template.FuncMap{
				"div": func(a, b int) int { return a / b },
			}).
			ParseFiles("web/templates/index.html"),
	)
	queueTmpl = template.Must(template.ParseFiles("web/templates/queue.html"))
)

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
	{"5m", "5m"},
	{"1h", "1h"},
	{"6h", "6h"},
	{"24h", "24h"},
	{"7d", "7d"},
}

var rangeDurations = map[string]time.Duration{
	"5m":  5 * time.Minute,
	"1h":  time.Hour,
	"6h":  6 * time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
}

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

	data := pageData{
		Vhosts:   vhosts,
		Selected: selected,
		Limits:   scraper.DefaultLimits,
	}

	if selected != "" {
		metrics, err := scraper.GetMetrics(selected)
		if err != nil {
			log.Printf("error fetching metrics for %s: %v", selected, err)
		} else {
			data.Metrics = metrics
		}
	}

	if err := indexTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

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

	samples, err := prom.QueryRange(prom.RangeOptions{
		Vhost: vhost,
		Queue: queue,
		Since: since,
		Step:  prom.StepFor(since),
	})
	if err != nil {
		log.Printf("prometheus error: %v", err)
	}

	samplesJSON := "[]"
	if len(samples) > 0 {
		b, _ := json.Marshal(samples)
		samplesJSON = string(b)
	}

	data := struct {
		Vhost         string
		Queue         string
		Ranges        []rangeOption
		SelectedRange string
		SamplesJSON   template.JS
		NoData        bool
	}{
		Vhost:         vhost,
		Queue:         queue,
		Ranges:        timeRanges,
		SelectedRange: rangeStr,
		SamplesJSON:   template.JS(samplesJSON),
		NoData:        len(samples) == 0,
	}

	if err := queueTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

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

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/queue", queueHandler)
	http.HandleFunc("/api/queues", queuesAPIHandler)
	http.HandleFunc("/api/cluster", clusterAPIHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Dashboard running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
