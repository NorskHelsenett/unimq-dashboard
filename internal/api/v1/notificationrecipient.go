package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get a notification recipient
// @Description	Get a specific notification recipient for a vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name		path		string	true	"Vhost Name"
// @Param			recipient-id	path		string	true	"Recipient ID"
// @Success		200				{object}	models.Recipient
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		401				{object}	httpsuite.ErrorResponse
// @Failure		403				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/recipients/{recipient-id} [get]
// @security		bearer
func (rc *APIService) GetNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required vhost parameter"),
		)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to decode vhost name"),
			httpsuite.WithInternalErrorMessage("error decoding vhost name: "+vhost),
		)
		return
	}

	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required recipient id parameter"),
		)
		return
	}

	recipient, err := rc.DB.GetNotificationRecipient(r.Context(), eVhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to fetch recipients"),
			httpsuite.WithInternalErrorMessage("failed to fetch recipients for vhost: "+eVhost),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &recipient)
	return

}

// @Summary		Add a new notification recipient
// @Description	Add a new notification recipient for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			recipient	body		models.PostRecipient	true	"Notification Recipient Object"
// @Success		201			{object}	string					"Recipient added successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/recipients [post]
// @security		bearer
func (rc *APIService) AddNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required vhost parameter"),
		)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to decode vhost name"),
			httpsuite.WithInternalErrorMessage("error decoding vhost name: "+vhost),
		)
		return
	}

	var recipient models.PostRecipient
	err = httpsuite.ReadResponse(r, &recipient)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}
	out, err := recipient.ToRecipient()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid recipient data"),
		)
		return
	}

	vhostNotification, err := rc.ensureNotificationHostExists(r.Context(), eVhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhost"),
		)
		return
	}

	err = rc.DB.AddNotificationRecipient(r.Context(), vhostNotification.Name, out)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to add recipient"),
		)
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
// @Failure		400				{object}	httpsuite.ErrorResponse
// @Failure		401				{object}	httpsuite.ErrorResponse
// @Failure		403				{object}	httpsuite.ErrorResponse
// @Failure		500				{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/recipients/{recipient-id} [delete]
// @security		bearer
func (rc *APIService) DeleteNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required vhost parameter"),
		)
		return
	}

	eVhost, err := url.QueryUnescape(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to decode vhost name"),
			httpsuite.WithInternalErrorMessage("error decoding vhost name: "+vhost),
		)
		return
	}

	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("missing required recipient id parameter"),
		)
		return
	}

	err = rc.DB.DeleteNotificationRecipient(r.Context(), eVhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to delete recipient"),
			httpsuite.WithInternalErrorMessage("failed to delete recipient: "+id),
		)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Recipient deleted successfully", http.StatusOK)
}
