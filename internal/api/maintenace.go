package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// @Summary		Get scheduled  maintenance information and history
// @Description	Get scheduled  maintenance information and history
// @Tags			Maintenance
// @Produce		json
// @Success		200	{object}	models.MaintenanceResponse
// @Failure		500	{object}	httpsuite.APIError
// @Router			/v1/maintenance [get]
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

	response := models.NewMaintenanceResponse(scheduled, maintenanceHistory)
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &response)
}

// @Summary		Get all maintenance entries
// @Description	Get all maintenance entries for admin view
// @Tags			Maintenance
// @Produce		json
// @Success		200	{object}	models.MaintenanceAdminResponse
// @Failure		500	{object}	httpsuite.APIError
// @Router			/v1/maintenance/admin [get]
func (rc *APIService) GetMaintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {

	maintenanceAll, err := rc.DB.GetMaintenanceAll(r.Context(), bson.M{})
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance", http.StatusInternalServerError)
		return
	}

	response := models.NewMaintenanceAdminResponse(maintenanceAll)
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &response)
}

// @Summary		Add new maintenance entry
// @Description	Add new maintenance entry with description, start time, and end time
// @Tags			Maintenance
// @Produce		json
// @Param			entry	body		models.MaintenanceEntry	true	"Maintenance Entry Data"
// @Success		303		{string}	string					"Redirect to maintenance admin page with success message"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/maintenance [post]
func (rc *APIService) AddMaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	var entry models.MaintenanceEntry
	err := httpsuite.ReadResponse(r, &entry)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if entry.Description == "" || entry.Start.IsZero() || entry.End.IsZero() {
		httpsuite.WriteJSONError(w, "description, start time, and end time are required", http.StatusBadRequest)
		return
	}

	rc.DB.AddMaintenanceEntry(r.Context(), &entry)

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

// @Summary		Update maintenance entry status
// @Description	Update the status of a maintenance entry (e.g., scheduled, in-progress, completed)
// @Tags			Maintenance
// @Accept			json
// @Produce		json
// @Param			maintenance	path		string						true	"Maintenance Entry ID"
// @Param			status		body		models.UpdateMaintenance	true	"Updated Maintenance Status"
// @Success		303			{string}	string						"Redirect to maintenance admin page with success message"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/maintenance/{maintenance}/status [put]
func (rc *APIService) UpdateMaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
		return
	}

	var request models.UpdateMaintenance
	err := httpsuite.ReadResponse(r, &request)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = rc.DB.SetMaintenanceEntryStatus(r.Context(), id, string(request.Status))
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating maintenance status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}

// @Summary		Delete a maintenance entry
// @Description	Delete a specific maintenance entry by ID
// @Tags			Maintenance
// @Param			maintenance	path		string	true	"Maintenance Entry ID"
// @Success		303			{string}	string	"Redirect to maintenance admin page with success message"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/maintenance/{maintenance} [delete]
func (rc *APIService) DeleteMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteMaintenanceEntry(r.Context(), id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting maintenance entry", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/maintenance/admin", http.StatusSeeOther)
}
