package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Add a new notification recipient
// @Description	Add a new notification recipient for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost		path		string				true	"Vhost Name"
// @Param			recipient	body		models.Recipient	true	"Notification Recipient Object"
// @Success		303			{string}	string				"Redirect to notifications page"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/recipients [post]
func (rc *APIService) AddNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}

	var recipient models.Recipient
	err := httpsuite.ReadResponse(r, &recipient)
	if err != nil {
		httpsuite.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	vhostObject, err := rc.DB.GetVhost(r.Context(), vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id := strconv.FormatInt(int64(len(vhostObject.Recipients)+1), 10)
	recipient.ID = id

	err = rc.DB.AddNotificationRecipient(r.Context(), vhost, recipient)
	if err != nil {
		httpsuite.WriteJSONError(w, "error adding recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}

// @Summary		Delete a notification recipient
// @Description	Delete a specific notification recipient for a vhost
// @Tags			Notifications
// @Param			vhost		path		string	true	"Vhost Name"
// @Param			recipient	path		string	true	"Recipient ID"
// @Success		303			{string}	string	"Redirect to notifications page"
// @Failure		400			{object}	httpsuite.APIError
// @Failure		500			{object}	httpsuite.APIError
// @Router			/v1/notifications/{vhost-name}/recipients/{recipient-id} [delete]
func (rc *APIService) DeleteNotificationsRecipientHandler(w http.ResponseWriter, r *http.Request) {
	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing required vhost parameter", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "recipient")
	if id == "" {
		httpsuite.WriteJSONError(w, "missing required id parameter", http.StatusBadRequest)
		return
	}

	err := rc.DB.DeleteNotificationRecipient(r.Context(), vhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w, "error deleting recipient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notifications?vhost="+url.QueryEscape(vhost), http.StatusSeeOther)
}
