package rmq

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

// @Summary		Get all vhosts
// @Description	Get a list of all vhosts in the RabbitMQ cluster
// @Tags			Vhosts
// @Produce		json
// @Success		200	{array}		[]models.Vhost
// @Failure		400	{object}	httpsuite.ErrorResponse
// @Failure		401	{object}	httpsuite.ErrorResponse
// @Failure		403	{object}	httpsuite.ErrorResponse
// @Failure		502	{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts [get]
// @security		bearer
func (rc *RMQHandler) GetVhostsHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhosts, err := rc.RMQClient.GetVhosts()
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhosts"),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhosts)
}

// @Summary		Get vhost details
// @Description	Get details of a specific vhost by name
// @Tags			Vhosts
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.Vhost
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts/{vhost-name} [get]
// @security		bearer
func (rc *RMQHandler) GetVhostHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("vhost name is required"),
		)
		return
	}

	eVhostName, err := url.QueryUnescape(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to decode vhost name"),
		)
		return
	}

	vhostData, err := rc.RMQClient.GetVhost(eVhostName)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhosts data"),
			httpsuite.WithInternalErrorMessage(fmt.Sprintf("failed to fetch vhost data for vhost '%s': %v", eVhostName, err)),
		)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhostData)
}

// @Summary		Get vhost limits
// @Description	Get limits of a specific vhost by name
// @Tags			Vhosts
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.RMQVhostLimits
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts/{vhost-name}/limits [get]
// @security		bearer
func (rc *RMQHandler) GetVhostLimitsHandler(w http.ResponseWriter, r *http.Request) {

	_, err := httpsuite.IsAGroupInClaim(r.Context(), rc.AdminGroups)
	if err != nil {
		httpsuite.WriteJSONErrorForbidden(w, httpsuite.WithInternalErrorMessage(err.Error()))
		return
	}

	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithErrorMessage("vhost name is required"),
		)
		return
	}

	eVhostName, err := url.QueryUnescape(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusBadRequest,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to decode vhost name"),
		)
		return
	}

	vhostLimits, err := rc.RMQClient.GetVhostLimit(eVhostName)
	if err != nil {
		if errors.Is(err, rabbitmq.ErrLimitsNotFound) {
			vhostLimits = models.NewRMQVhostLimits(eVhostName)
		} else {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to fetch vhost limits"),
				httpsuite.WithInternalErrorMessage(fmt.Sprintf("failed to fetch vhost limits for vhost '%s': %v", eVhostName, err)),
			)
			return
		}
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhostLimits)

}

// @Summary		Get Vhost usage
// @Description	Get usage statistics for a specific vhost in the RabbitMQ cluster
// @Tags			RabbitMQ
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	[]models.RMQVhostUsage
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhost/{vhost-name}/usage [get]
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
