package api

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @Summary		Get a notification rule
// @Description	Retrieve a specific notification rule for a vhost
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string				true	"Vhost Name"
// @Param			rule-id		path		string				true	"Notification Rule ID"
// @Success		200			{object}	models.AlarmRule	"Notification rule retrieved successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		404			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) GetNotificationRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}

	rule, err := rc.DB.GetNotificationRule(r.Context(), eVhost, id)
	if err != nil {
		if errors.Is(err, database.ErrNotificationRuleNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("notification rule not found"),
			)
			return
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("notification rule not found"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch notification rule"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Notification rule retrieved successfully", http.StatusOK, rule)
}

// @Summary		Add a new notification rule
// @Description	Add a new notification rule for a specific vhost
// @Tags			Notifications
// @Accept			json
// @Produce		json
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			rule		body		models.PostAlarmRule	true	"Notification Rule Object"
// @Success		201			{object}	string					"Rule added successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		404			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules [post]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) AddNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	var rule models.PostAlarmRule
	err = httpsuite.ReadResponse(w, r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}

	_, err = rc.ensureNotificationHostExists(r.Context(), eVhost)
	if err != nil {
		if errors.Is(err, rabbitmq.ErrVhostNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("vhost not found"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("error validating vhost"),
		)
		return
	}

	out, err := rule.ToAlarmRule()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to convert rule data"),
		)
		return
	}

	err = rc.DB.AddNotificationRule(r.Context(), eVhost, out)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to add notification rule"),
		)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Rule added successfully", http.StatusCreated)
}

// @Summary		Delete a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Param			rule-id		path		string	true	"Notification Rule ID"
// @Success		200			{string}	string	"Rule deleted successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [delete]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) DeleteNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}
	err = rc.DB.DeleteNotificationRule(r.Context(), eVhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to delete notification rule"),
		)
		return
	}

	httpsuite.SendEmptyResponse(r.Context(), w, "Rule deleted successfully", http.StatusOK)
}

// @Summary		Update a notification rule
// @Description	Delete a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string					true	"Vhost Name"
// @Param			rule-id		path		string					true	"Notification Rule ID"
// @Param			rule		body		models.AlarmRuleUpdate	true	"Updated Notification Rule Object"
// @Success		200			{string}	string					"Rule updated successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [Post]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) UpdateNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}

	var rule models.AlarmRuleUpdate
	err = httpsuite.ReadResponse(w, r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}

	if rule.Threshold != nil {
		err = rc.DB.UpdateNotificationRuleThreshold(r.Context(), eVhost, id, *rule.Threshold)
		if err != nil {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to update rule threshold"),
			)
			return
		}
	}

	if rule.Message != nil {
		err = rc.DB.UpdateNotificationRuleMessage(r.Context(), eVhost, id, *rule.Message)
		if err != nil {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to update rule message"),
			)
			return
		}
	}

	httpsuite.SendResponse(r.Context(), w, "Rule updated successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Toggle a notification rule
// @Description	Enable or disable a specific notification rule for a vhost
// @Tags			Notifications
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Param			rule-id		path		string	true	"Notification Rule ID"
// @Success		200			{string}	string	"Rule toggled successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/toggle [post]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) ToggleNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}

	rule, err := rc.DB.GetNotificationRule(r.Context(), eVhost, id)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch notification rule"),
		)
		return
	}

	if rule == nil {
		httpsuite.WriteJSONError(w,
			http.StatusNotFound,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("notification rule not found"),
		)
		return
	}

	err = rc.DB.ToggleNotificationRule(r.Context(), eVhost, id, !rule.Enabled)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithExternalErrorMessage("failed to toggle notification rule"),
			httpsuite.WithInternalErrorMessage("failed to toggle notification rule: "+rule.ID),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "Rule toggled successfully", http.StatusOK, httpsuite.NewEmptyResponse())
}

// @Summary		Test Notification Rule
// @Description	Send a test notification using the specified rule to verify its configuration
// @Tags			Notifications
// @Produce		json
// @Param			vhost-name	path		string							true	"Vhost Name"
// @Param			rule-id		path		string							true	"Notification Rule ID"
// @Success		200			{object}	models.TestNotificationResponse	"Test notification sent successfully"
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/test [post]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *APIService) TestNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost-name")
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

	id := chi.URLParam(r, "rule-id")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule parameter"),
		)
		return
	}

	rule, err := rc.DB.GetNotificationRule(r.Context(), eVhost, id)
	if err != nil {
		if errors.Is(err, database.ErrVhostNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("vhost not found"),
			)
			return
		}
		if errors.Is(err, database.ErrNotificationRuleNotFound) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("notification rule not found"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch notification rule"),
		)
		return
	}

	vhostobject, err := rc.DB.GetVhost(r.Context(), eVhost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpsuite.WriteJSONError(w,
				http.StatusNotFound,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("vhost not found"),
			)
			return
		}
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhost"),
		)
		return
	}

	subject := "[UniMQ TEST] " + rule.Name + " — " + eVhost
	body := "This is a test message from UniMQ.\n\n" + rule.BuildMessage(eVhost)

	urls := vhostobject.WebhookURLs()
	if len(urls) == 0 {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("no webhook URLs configured for vhost "+eVhost),
		)
		return
	}
	notificationStatus := notificationhelper.NewNotifyStatus(urls, vhostobject.EmailRecipients())

	notificationStatus.WebhookStatuses = notificationhelper.SendWebhooks(urls, subject, body)
	if err != nil {
		slog.ErrorContext(r.Context(), "error sending test notification", "error", err)
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to send test notification"),
		)
		return
	}

	notificationStatus.EmailStatuses = notificationhelper.EmailSenderInstance.SendEmails(r.Context(), vhostobject.EmailRecipients(), subject, body, "text/plain")
	if err != nil {
		if errors.Is(err, notificationhelper.ErrEmailNotConfigured) {
			slog.WarnContext(r.Context(), "test email not sent, SMTP server is not configured", "emails", vhostobject.EmailRecipients())
		} else {
			slog.ErrorContext(r.Context(), "test email failed on some", "error", err)
		}
	} else {
		slog.InfoContext(r.Context(), "test email sent", "emails", vhostobject.EmailRecipients())
	}

	switch {
	case notificationStatus.IsTotalSuccess():
		response := models.TestNotificationResponse{
			Success:            true,
			Message:            "Test notification sent!",
			FailedDestinations: []string{},
		}
		httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
	case notificationStatus.IsPartialFailure():
		response := models.TestNotificationResponse{
			Success:            false,
			Message:            "Test notification sent with some failures.",
			FailedDestinations: notificationStatus.FailedDestinations(),
		}
		httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
	case notificationStatus.IsTotalFailure():
		response := models.TestNotificationResponse{
			Success:            false,
			Message:            "Test notification failed to send.",
			FailedDestinations: notificationStatus.FailedDestinations(),
		}
		httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
	default:
		response := models.TestNotificationResponse{
			Success:            false,
			Message:            "Test notification status unknown.",
			FailedDestinations: []string{},
		}
		httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusInternalServerError, &response)
	}

}
