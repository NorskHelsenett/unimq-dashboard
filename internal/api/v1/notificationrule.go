package api

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/wneessen/go-mail"
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
// @Failure		404			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [get]
// @security		bearer
func (rc *APIService) GetNotificationRuleHandler(w http.ResponseWriter, r *http.Request) {
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
// @Failure		404			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules [post]
// @security		bearer
func (rc *APIService) AddNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	var rule models.PostAlarmRule
	err = httpsuite.ReadResponse(r, &rule)
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
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [delete]
// @security		bearer
func (rc *APIService) DeleteNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id} [Post]
// @security		bearer
func (rc *APIService) UpdateNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("missing required rule id parameter"),
		)
		return
	}

	var rule models.AlarmRuleUpdate
	err = httpsuite.ReadResponse(r, &rule)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("invalid request body"),
		)
		return
	}

	err = rc.DB.UpdateNotificationRuleThreshold(r.Context(), eVhost, id, rule.Threshold)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to update rule threshold"),
		)
		return
	}

	err = rc.DB.UpdateNotificationRuleMessage(r.Context(), eVhost, id, rule.Message)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to update rule message"),
		)
		return
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
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/toggle [post]
// @security		bearer
func (rc *APIService) ToggleNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/notifications/{vhost-name}/rules/{rule-id}/test [post]
// @security		bearer
func (rc *APIService) TestNotificationsRuleHandler(w http.ResponseWriter, r *http.Request) {
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

	id := chi.URLParam(r, "rule")
	if id == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
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
		if err == mongo.ErrNoDocuments {
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
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("no webhook URLs configured for vhost "+eVhost),
		)
		return
	}

	err = notificationhelper.SendWebhooks(urls, subject, body)
	if err != nil {
		slog.ErrorContext(r.Context(), "error sending test notification", "error", err)
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to send test notification"),
		)
		return
	}

	if rc.EmailClient != nil {
		emails := vhostobject.EmailRecipients()
		for _, email := range emails {
			err = notificationhelper.SendEmail(rc.EmailConfig, email, subject, body, mail.TypeTextPlain)
			if err != nil {
				slog.ErrorContext(r.Context(), "error sending test email", "error", err)
				httpsuite.WriteJSONError(w,
					http.StatusBadRequest,
					httpsuite.WithError(err),
					httpsuite.WithErrorMessage("failed to send test email"),
				)
				return
			}
		}
	} else {
		slog.WarnContext(r.Context(), "email client not configured, skipping email test notification")
	}

	response := models.TestNotificationResponse{
		Success: true,
		Message: "Test notification sent!",
	}

	httpsuite.SendResponse(r.Context(), w, "Testing notification...", http.StatusOK, &response)
}
