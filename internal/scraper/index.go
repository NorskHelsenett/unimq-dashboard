package scraper

import (
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) IndexHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, err := rc.GetVhosts()
	if err != nil {
		http.Error(w, "Could not reach RabbitMQ: "+err.Error(), http.StatusBadGateway)
		return
	}
	selected := r.URL.Query().Get("vhost")
	if selected == "" && len(vhosts) > 0 {
		selected = vhosts[0]
	}
	data := models.PageData{Vhosts: vhosts, Selected: selected, Limits: *rc.RMQLimits}
	if selected != "" {
		if m, err := rc.GetMetrics(selected); err == nil {
			data.Metrics = m
		}
	}
	if err := templating.IndexTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}
