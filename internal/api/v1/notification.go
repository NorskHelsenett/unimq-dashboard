package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary		Get Notifications
// @Description	Get all notification vhosts and settings
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	[]models.VhostNotification
// @Failure		502	{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications [get]
// @security		bearer
func (rc *APIService) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	notifications, err := rc.DB.GetNotificationsAll(r.Context())
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch notifications"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notifications", http.StatusOK, &notifications)
}

// @Summary		Get Notification on Vhost
// @Description	Get notification rules and settings for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.VhostNotification
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name} [get]
// @security		bearer
func (rc *APIService) GetNotificationsVhostHandler(w http.ResponseWriter, r *http.Request) {

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

	notification, err := rc.DB.GetNotification(r.Context(), eVhost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			empty := models.NewVhostNotification(eVhost)
			httpsuite.SendResponse(r.Context(), w, "Gathered notifications on vhost", http.StatusOK, empty)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch notifications"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Gathered notifications on vhost", http.StatusOK, notification)
}

// @Summary		Delete Notifications
// @Description	Deletes the notification configuration for a specific vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	string
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name} [delete]
// @security		bearer
func (rc *APIService) DeleteNotificationsHandler(w http.ResponseWriter, r *http.Request) {
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

	err = rc.DB.DeleteNotification(r.Context(), eVhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to delete notifications"),
		)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Deleted notifications on vhost", http.StatusOK)
}
