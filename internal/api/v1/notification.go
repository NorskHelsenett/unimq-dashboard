package api

import (
	"errors"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/requesthelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary		Get Notifications
// @Description	Get all notification vhosts and settings
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	[]models.VhostNotification
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		502	{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

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
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name} [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetNotificationsVhostHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost, err := requesthelper.ReadVhostFromRequest(r)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to read vhost parameter"),
		)
		return
	}

	notification, err := rc.DB.GetNotification(r.Context(), vhost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			empty := models.NewVhostNotification(vhost)
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
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name} [delete]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) DeleteNotificationsHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost, err := requesthelper.ReadVhostFromRequest(r)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to read vhost parameter"),
		)
		return
	}

	err = rc.DB.DeleteNotification(r.Context(), vhost)
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
