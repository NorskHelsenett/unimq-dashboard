package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type maintenanceData struct {
	Scheduled []models.MaintenanceEntry
	History   []models.MaintenanceEntry
}

// @Summary Get scheduled  maintenance information and history
// @Description Get scheduled  maintenance information and history
// @Tags Maintenance
// @Produce json
// @Success 200 {object} maintenanceData
// @Failure 500 {object} httpsuite.ErrorResponse
// @Router /maintenance [get]
func (rc *APIService) GetMaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	scheduled, err := rc.DB.GetMaintenanceScheduled(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching scheduled maintenance", http.StatusInternalServerError)
		return
	}

	maintenanceHistory, err := rc.DB.GetMaintenanceHistory(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance history", http.StatusInternalServerError)
		maintenanceHistory = []models.MaintenanceEntry{}
	}

	data := maintenanceData{
		Scheduled: scheduled,
		History:   maintenanceHistory,
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &data)
}

type maintenanceAdminData struct {
	Entries []models.MaintenanceEntry
}

// @Summary Get all maintenance entries
// @Description Get all maintenance entries for admin view
// @Tags Maintenance
// @Produce json
// @Success 200 {object} maintenanceAdminData
// @Failure 500 {object} httpsuite.ErrorResponse
// @Router /maintenance/admin [get]
func (rc *APIService) GetMaintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {

	maintenanceAll, err := rc.DB.GetMaintenanceAll(r.Context(), bson.M{})
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance", http.StatusInternalServerError)
		return
	}
	adminData := maintenanceAdminData{
		Entries: maintenanceAll,
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &adminData)
}

func (rc *APIService) AddMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
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
		Status:      models.MaintenanceStatusScheduled,
	}
	rc.DB.AddMaintenanceEntry(r.Context(), &entry)

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *APIService) UpdateMaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		httpsuite.WriteJSONError(w, "status is required", http.StatusBadRequest)
		return
	}

	if models.IsValidMaintenanceStatus(status) == false {
		httpsuite.WriteJSONError(w, "invalid status value, expected any of "+strings.Join(models.GetMaintenanceStatusAllString(), ", "), http.StatusBadRequest)
		return
	}

	err := rc.DB.SetMaintenanceEntryStatus(r.Context(), id, status)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating maintenance status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

func (rc *APIService) DeleteMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
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
