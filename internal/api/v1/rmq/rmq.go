package rmq

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
)

type RMQHandler struct {
	RMQClient   *rabbitmq.RMQClient
	AdminGroups []string
}

func NewRMQHandler(rmqClient *rabbitmq.RMQClient, adminGroups []string) *RMQHandler {
	return &RMQHandler{
		RMQClient:   rmqClient,
		AdminGroups: adminGroups,
	}
}

// @Summary		Get node statistics
// @Description	Get statistics for all nodes in the RabbitMQ cluster
// @Tags			RabbitMQ
// @Produce		json
// @Success		200	{object}	[]models.RMQNode
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/rabbitmq [get]
// @security		bearer
func (rc *RMQHandler) GetRMQNodesHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	stats, err := rc.RMQClient.GetNodes()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch cluster stats"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered cluster stats", http.StatusOK, &stats)
}

// @Summary		Get Vhost usage
// @Description	Get usage statistics for all vhosts in the RabbitMQ cluster
// @Tags			RabbitMQ
// @Produce		json
// @Success		200	{object}	[]models.RMQVhostUsage
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/rabbitmq/vhostusage [get]
// @security		bearer
func (rc *RMQHandler) GetRMQVhostUsageHandler(w http.ResponseWriter, r *http.Request) {
	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("vhost name is required"),
		)
		return
	}

	usage, err := rc.RMQClient.GetVhostUsage(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhost usage"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered vhost usage", http.StatusOK, &usage)
}

// @Summary		Get Vhost limits
// @Description	Get limits for all vhosts in the RabbitMQ cluster
// @Tags			RabbitMQ
// @Produce		json
// @Success		200	{object}	[]models.RMQLimits
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		500	{object}	httpsuite.ErrorResponse
// @Router			/v1/rabbitmq/vhostlimits [get]
// @security		bearer
func (rc *RMQHandler) GetRMQLimitsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	limits, err := rc.RMQClient.GetLimits()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch cluster limits"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered cluster limits", http.StatusOK, &limits)
}

// @Summary		Get Vhost limits
// @Description	Get limits for a specific vhost in the RabbitMQ cluster
// @Tags			RabbitMQ
// @Produce		json
// @Param			vhost	path		string	true	"Vhost Name"
// @Success		200		{object}	models.RMQLimits
// @Failure		400		{object}	httpsuite.ErrorResponse
// @Failure		401		{object}	httpsuite.ErrorResponse
// @Failure		403		{object}	httpsuite.ErrorResponse
// @Failure		404		{object}	httpsuite.ErrorResponse
// @Failure		500		{object}	httpsuite.ErrorResponse
// @Router			/v1/rabbitmq/vhostlimits/{vhost} [get]
// @security		bearer
func (rc *RMQHandler) GetRMQVhostLimitsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhost := chi.URLParam(r, "vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("vhost name is required"),
		)
		return
	}

	limits, err := rc.RMQClient.GetLimit(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhost limits"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "gathered vhost limits", http.StatusOK, &limits)
}
