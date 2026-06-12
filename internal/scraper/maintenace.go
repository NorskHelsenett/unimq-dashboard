package scraper

import (
	"log"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type maintenanceData struct {
	Scheduled []models.MaintenanceEntry
	History   []models.MaintenanceEntry
}

func (rc *RestClient) GetMaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	scheduled, err := rc.DB.GetMaintenanceScheduled(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching scheduled maintenance", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error fetching scheduled maintenance", "error", err)
		return
	}

	maintenanceHistory, err := rc.DB.GetMaintenanceHistory(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance history", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error fetching maintenance history", "error", err)
		maintenanceHistory = []models.MaintenanceEntry{}
	}

	data := maintenanceData{
		Scheduled: scheduled,
		History:   maintenanceHistory,
	}

	if err := templating.MaintTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

type maintenanceAdminData struct {
	Entries []models.MaintenanceEntry
}

func (rc *RestClient) GetMaintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {

	maintenanceAll, err := rc.DB.GetMaintenanceAll(r.Context(), bson.M{})
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error fetching maintenance", "error", err)
		return
	}
	adminData := maintenanceAdminData{
		Entries: maintenanceAll,
	}

	if err := templating.MaintAdminTmpl.Execute(w, adminData); err != nil {
		httpsuite.WriteJSONError(w, "error rendering maintenance admin template", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error rendering maintenance admin template", "error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rc *RestClient) PostMaintenanceAddHandler(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	if description == "" {
		httpsuite.WriteJSONError(w, "description is required", http.StatusBadRequest)
		return
	}

	start := r.FormValue("start")
	if start == "" {
		httpsuite.WriteJSONError(w, "start time is required", http.StatusBadRequest)
		return
	}

	end := r.FormValue("end")
	if end == "" {
		httpsuite.WriteJSONError(w, "end time is required", http.StatusBadRequest)
		return
	}

	startTime, err := time.ParseInLocation("2006-01-02T15:04", start, time.Local)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid start time format", http.StatusBadRequest)
		return
	}
	endTime, err := time.ParseInLocation("2006-01-02T15:04", end, time.Local)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid end time format", http.StatusBadRequest)
		return
	}

	entry := models.MaintenanceEntry{
		Description: description,
		Start:       startTime,
		End:         endTime,
		Status:      "scheduled",
	}
	rc.DB.AddMaintenanceEntry(r.Context(), &entry)

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *RestClient) PostMaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		httpsuite.WriteJSONError(w, "status is required", http.StatusBadRequest)
		return
	}

	err := rc.DB.SetMaintenanceEntryStatus(r.Context(), id, status)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating maintenance status", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error updating maintenance status", "error", err)
		return
	}

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *RestClient) PostMaintenanceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		httpsuite.WriteJSONError(w, "id is required", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteMaintenanceEntry(r.Context(), id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting maintenance entry", http.StatusInternalServerError)
		// slog.ErrorContext(r.Context(), "error deleting maintenance entry", "error", err)
		return
	}

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}
