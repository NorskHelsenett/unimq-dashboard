package scraper

import (
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) IndexHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, err := rc.RMQClient.GetVhosts()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhosts", http.StatusBadGateway)
		return
	}
	selected := r.URL.Query().Get("vhost")
	if selected == "" && len(vhosts) > 0 {
		selected = vhosts[0]
	}
	data := models.PageData{Vhosts: vhosts, Selected: selected, Limits: *rc.RMQLimits}
	if selected != "" {
		metrics, err := rc.RMQClient.GetMetrics(selected)
		if err == nil {
			data.Metrics = metrics
		}
	}
	if err := templating.IndexTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}
