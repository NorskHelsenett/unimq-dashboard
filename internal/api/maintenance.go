package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
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
	httpsuite.SendResponse(r.Context(), w, "Fetched sheduled and historic maintenance", http.StatusOK, &response)
}

// @Summary		Get all maintenance entries
// @Description	Get all maintenance entries for admin view
// @Tags			Maintenance
// @Produce		json
// @Success		200	{object}	models.MaintenanceAdminResponse
// @Failure		500	{object}	httpsuite.APIError
// @Router			/v1/maintenance/admin [get]
func (rc *APIService) GetMaintenanceAdminHandler(w http.ResponseWriter, r *http.Request) {

	maintenanceAll, err := rc.DB.GetMaintenanceAll(r.Context(), bson.D{})
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching maintenance", http.StatusInternalServerError)
		return
	}

	response := models.NewMaintenanceAdminResponse(maintenanceAll)
	httpsuite.SendResponse(r.Context(), w, "Fetched all maintenance entries", http.StatusOK, &response)
}

// @Summary		Add new scheduled maintenance entry
// @Description	Add new maintenance entry with description, start time, and end time that will have the status Scheduled
// @Tags			Maintenance
// @Produce		json
// @Param			entry	body		models.PostMaintenanceEntry	true	"Maintenance Entry Data"
// @Success		201		{string}	string						"Maintenance entry added successfully"
// @Failure		400		{object}	httpsuite.APIError
// @Failure		500		{object}	httpsuite.APIError
// @Router			/v1/maintenance [post]
func (rc *APIService) AddMaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	var entry models.PostMaintenanceEntry
	err := httpsuite.ReadResponse(r, &entry)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	dbentry, err := entry.ToMaintenanceEntry()
	if err != nil {
		slog.Error("error converting to maintenance entry", "error", err)
		httpsuite.WriteJSONError(w, "invalid maintenance entry data: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = rc.DB.AddMaintenanceEntry(r.Context(), dbentry)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding maintenance entry", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance entry added successfully", http.StatusCreated, httpsuite.NewEmptyResponse())
}

// @Summary		Update maintenance entry status
// @Description	Update the status of a maintenance entry (e.g., scheduled, in-progress, completed)
// @Tags			Maintenance
// @Accept			json
// @Produce		json
// @Param			maintenance-id	path		string						true	"Maintenance Entry ID"
// @Param			status			body		models.UpdateMaintenance	true	"New Maintenance Status"
// @Success		200				{string}	string						"Maintenance status updated successfully"
// @Failure		400				{object}	httpsuite.APIError
// @Failure		404				{object}	httpsuite.APIError
// @Failure		500				{object}	httpsuite.APIError
// @Router			/v1/maintenance/{maintenance-id} [put]
func (rc *APIService) UpdateMaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
		return
	}

	var request models.UpdateMaintenance
	err := httpsuite.ReadResponse(r, &request)
	if err != nil {
		httpsuite.WriteJSONError(w, fmt.Sprintf("invalid request body %w. expected any of %v"+err.Error(), models.GetMaintenanceStatusAllString()), http.StatusBadRequest)
		return
	}

	err = rc.DB.SetMaintenanceEntryStatus(r.Context(), id, request.Status)
	if err != nil {
		httpsuite.WriteJSONError(w, "error updating maintenance status", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance status updated successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Delete a maintenance entry
// @Description	Delete a specific maintenance entry by ID
// @Tags			Maintenance
// @Param			maintenance-id	path	string	true	"Maintenance Entry ID"
// @Produce		json
// @Success		200	{string}	string	"Maintenance entry deleted successfully"
// @Failure		400	{object}	httpsuite.APIError
// @Failure		404	{object}	httpsuite.APIError
// @Failure		500	{object}	httpsuite.APIError
// @Router			/v1/maintenance/{maintenance-id} [delete]
func (rc *APIService) DeleteMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w, "maintenance id is required", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteMaintenanceEntry(r.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrMaintenanceNotFound) {
			httpsuite.WriteJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		httpsuite.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance entry deleted successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}
