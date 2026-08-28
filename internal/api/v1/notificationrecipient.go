package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Add a new notification recipient
// @Description	Add a new notification recipient for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			recipient	body		models.PostRecipient	true	"Notification Recipient Object"
// @Success		201			{object}	string					"Recipient added successfully"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/recipients [post]
// @security		bearer
func (rc *APIService) AddNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	var recipient models.PostRecipient
	err = httpsuite.ReadResponse(r, &recipient)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	out, err := recipient.ToRecipient()
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid recipient data: "+err.Error(), http.StatusBadRequest)
		return
	}

	vhostNotification, err := rc.ensureNotificationHostExists(r.Context(), eVhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = rc.DB.AddNotificationRecipient(r.Context(), vhostNotification.Name, out)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Recipient added successfully", http.StatusCreated)
}

// @Summary		Delete a notification recipient
// @Description	Delete a specific notification recipient for a vhost
// @Tags			Notifications
// @Param			vhost-name		path		string	true	"Vhost Name"
// @Param			recipient-id	path		string	true	"Recipient ID"
// @Success		200				{string}	string	"Recipient deleted successfully"
// @Failure		400				{object}	httpsuite.APIError
// @Failure		500				{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/recipients/{recipient-id} [delete]
// @security		bearer
func (rc *APIService) DeleteNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	err = rc.DB.DeleteNotificationRecipient(r.Context(), eVhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Recipient deleted successfully", http.StatusOK)
}
