package api

import (
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

// @summary Dashboard index page
// @description Displays an overview of RabbitMQ vhosts and metrics.
// @tags dashboard
// @produce html
// @param vhost query string false "Optional vhost to select"
// @success 200 {string} string "HTML page"
// @failure 502 {object} httpsuite.ErrorResponse "Error fetching vhosts"
// @router / [get]
func (rc *APIService) IndexHandler(w http.ResponseWriter, r *http.Request) {
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
