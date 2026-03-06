package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

var tmpl = template.Must(
	template.New("index.html").
		Funcs(template.FuncMap{
			"div": func(a, b int) int { return a / b },
		}).
		ParseFiles("web/templates/index.html"),
)

type pageData struct {
	Vhosts   []string
	Selected string
	Metrics  *scraper.VhostMetrics
	Limits   scraper.Limits
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

	if err := tmpl.Execute(w, data); err != nil {
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

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/queues", queuesAPIHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Dashboard running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
