package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get scheduled  maintenance information and history
// @Description	Get scheduled  maintenance information and history
// @Tags			Maintenance
// @Produce		json
// @Success		200	{object}	models.MaintenanceResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance [get]
// @security		bearer
func (rc *APIService) GetMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	// Advance stale entries before returning so callers always see current statuses
	if _, err := rc.DB.AdvanceMaintenanceStatuses(r.Context()); err != nil {
		slog.WarnContext(r.Context(), "failed to advance maintenance statuses", "error", err)
	}

	scheduled, err := rc.DB.GetMaintenanceScheduled(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch scheduled maintenance"),
		)
		return
	}

	maintenanceHistory, err := rc.DB.GetMaintenanceHistory(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch maintenance history"),
		)
		maintenanceHistory = []models.MaintenanceEntry{}
	}

	response := models.NewMaintenanceResponse(scheduled, maintenanceHistory)
	httpsuite.SendResponse(r.Context(), w, "Fetched sheduled and historic maintenance", http.StatusOK, &response)
}

// @Summary		Get a specific maintenance entry
// @Description	Get a specific maintenance entry by ID
// @Tags			Maintenance
// @Produce		json
// @Param			maintenance-id	path		string	true	"Maintenance Entry ID"
// @Success		200				{object}	models.MaintenanceEntry
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		404				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance/{maintenance-id} [get]
// @security		bearer
func (rc *APIService) GetMaintenanceEntryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("maintenance id is required"),
		)
		return
	}

	entry, err := rc.DB.GetMaintenanceEntry(r.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrMaintenanceNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("maintenance entry not found"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch maintenance entry"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Fetched maintenance entry", http.StatusOK, &entry)
}

// @Summary		Add new scheduled maintenance entry
// @Description	Add new maintenance entry with description, start time, and end time that will have the status Scheduled
// @Tags			Maintenance
// @Produce		json
// @Param			entry	body		models.PostMaintenanceEntry	true	"Maintenance Entry Data"
// @Success		201		{string}	string						"Maintenance entry added successfully"
// @Failure		400		{object}	httpsuite.ErrorResponse
// @Failure		500		{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance [post]
// @security		bearer
func (rc *APIService) AddMaintenanceHandler(w http.ResponseWriter, r *http.Request) {

	var entry models.PostMaintenanceEntry
	err := httpsuite.ReadResponse(r, &entry)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}

	dbentry, err := entry.ToMaintenanceEntry()
	if err != nil {
		slog.Error("error converting to maintenance entry", "error", err)
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to convert to maintenance entry"),
		)
		return
	}

	err = rc.DB.AddMaintenanceEntry(r.Context(), dbentry)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to add maintenance entry"),
		)
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
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		404				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance/{maintenance-id} [put]
// @security		bearer
func (rc *APIService) UpdateMaintenanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("maintenance id is required"),
		)
		return
	}

	var request models.UpdateMaintenance
	err := httpsuite.ReadResponse(r, &request)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage(fmt.Sprintf("invalid request, expected any of %v", models.GetMaintenanceStatusAllString())),
		)
		return
	}

	err = rc.DB.SetMaintenanceEntryStatus(r.Context(), id, request.Status)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to update maintenance status"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance status updated successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Edit a maintenance entry
// @Description	Edit description, start, and end of an existing maintenance entry, with an audit trail
// @Tags			Maintenance
// @Accept			json
// @Produce		json
// @Param			maintenance-id	path		string							true	"Maintenance Entry ID"
// @Param			entry			body		models.PatchMaintenanceEntry	true	"Updated Maintenance Data"
// @Success		200				{string}	string							"Maintenance entry updated successfully"
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		404				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance/{maintenance-id} [patch]
func (rc *APIService) PatchMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("maintenance id is required"),
		)
		return
	}

	var request models.PatchMaintenanceEntry
	if err := httpsuite.ReadResponse(r, &request); err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}

	if err := request.Validate(); err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("request validation failed"),
		)
		return
	}

	start, _ := time.Parse("2006-01-02 15:04:05", request.Start)
	end, _ := time.Parse("2006-01-02 15:04:05", request.End)

	err := rc.DB.PatchMaintenanceEntry(r.Context(), id, request.Description, start, end, request.Reason, request.UpdatedBy)
	if err != nil {
		if errors.Is(err, database.ErrMaintenanceNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to find maintenance entry"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to update maintenance entry"),
		)
		return
	}

	logEntry := &models.MaintenanceEditLog{
		ID:            uuid.New().String(),
		MaintenanceID: id,
		Description:   request.Description,
		Start:         start,
		End:           end,
		Reason:        request.Reason,
		UpdatedBy:     request.UpdatedBy,
		UpdatedAt:     time.Now().UTC(),
	}
	if logErr := rc.DB.AddMaintenanceEditLog(r.Context(), logEntry); logErr != nil {
		slog.WarnContext(r.Context(), "failed to write maintenance edit log", "error", logErr)
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance entry updated successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Get edit history for a maintenance entry
// @Description	Returns all edit log entries for a given maintenance ID
// @Tags			Maintenance
// @Produce		json
// @Param			maintenance-id	path		string	true	"Maintenance Entry ID"
// @Success		200				{array}		models.MaintenanceEditLog
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance/{maintenance-id}/logs [get]
func (rc *APIService) GetMaintenanceEditLogsHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("maintenance id is required"),
		)
		return
	}

	logs, err := rc.DB.GetMaintenanceEditLogs(r.Context(), id)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch maintenance edit logs"),
		)
		return
	}

	type logsResponse struct {
		Logs []models.MaintenanceEditLog `json:"logs"`
	}
	httpsuite.SendResponse(r.Context(), w, "Fetched maintenance edit logs", http.StatusOK, &logsResponse{Logs: logs})
}

// @Summary		Delete a maintenance entry
// @Description	Delete a specific maintenance entry by ID
// @Tags			Maintenance
// @Param			maintenance-id	path	string	true	"Maintenance Entry ID"
// @Produce		json
// @Success		200	{string}	string	"Maintenance entry deleted successfully"
// @Failure		400	{object}	httpsuite.ErrorResponse
// @Failure		404	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/maintenance/{maintenance-id} [delete]
// @security		bearer
func (rc *APIService) DeleteMaintenanceHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "maintenance")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("maintenace id is required"),
		)
		return
	}

	err := rc.DB.DeleteMaintenanceEntry(r.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrMaintenanceNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Maintenance entry deleted successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}
