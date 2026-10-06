package rmq

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/requesthelper"
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
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
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
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) GetVhostHandler(w http.ResponseWriter, r *http.Request) {

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

	filter := rabbitmq.Filter{
		Parameter: rabbitmq.ParameterName,
		Value:     vhost,
	}
	vhostData, err := rc.RMQClient.GetVhosts(filter)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch vhosts data"),
			httpsuite.WithInternalErrorMessage(fmt.Sprintf("failed to fetch vhost data for vhost '%s': %v", vhost, err)),
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
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) GetVhostLimitsHandler(w http.ResponseWriter, r *http.Request) {

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

	filter := rabbitmq.NewFilter(rabbitmq.ParameterName, rabbitmq.FilterTypeVhost, vhost)
	vhostLimits, err := rc.RMQClient.GetVhostLimits(filter)
	if err != nil {
		// If the limits are not found, we return an empty limits object instead of an error as there are no limits set.
		if errors.Is(err, rabbitmq.ErrLimitsNotFound) {
			vhostLimits = models.NewRMQVhostLimits(vhost)
		} else {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to fetch vhost limits"),
				httpsuite.WithInternalErrorMessage(fmt.Sprintf("failed to fetch vhost limits for vhost '%s': %v", vhost, err)),
			)
			return
		}
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhostLimits)

}

// @Summary		Get Vhost usage
// @Description	Get usage statistics for a specific vhost in the RabbitMQ cluster
// @Tags			Vhosts
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	[]models.RMQVhostUsage
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		500			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts/{vhost-name}/usage [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) GetRMQVhostUsageHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := rabbitmq.NewFilter(rabbitmq.ParameterName, rabbitmq.FilterTypeVhost, vhost)
	usage, err := rc.RMQClient.GetVhostUsage(filter)
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
