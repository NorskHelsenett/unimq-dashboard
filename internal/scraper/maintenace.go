package scraper

import (
	"log"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) GetMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Scheduled []models.MaintenanceEntry
		History   []models.MaintenanceEntry
	}{templating.MaintStore.Scheduled(), templating.MaintStore.History()}
	if err := templating.MaintTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (rc *RestClient) GetMaintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {
	data := struct{ Entries []models.MaintenanceEntry }{templating.MaintStore.All()}
	if err := templating.MaintAdminTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (rc *RestClient) MaintenanceAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		templating.MaintStore.Add(r.FormValue("description"), r.FormValue("start"), r.FormValue("end"))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *RestClient) MaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		templating.MaintStore.SetStatus(r.FormValue("id"), models.ParseMaintenanceStatus(r.FormValue("status")))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *RestClient) MaintenanceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		templating.MaintStore.Delete(r.FormValue("id"))
	}
	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}
