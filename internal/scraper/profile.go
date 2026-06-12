package scraper

import (
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := templating.ProfileTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}
