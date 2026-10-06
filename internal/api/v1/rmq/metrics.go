package rmq

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/requesthelper"
)

// @Summary		Get Vhost Metrics
// @Description	Get real-time metrics for a specific vhost, including queue lengths, message rates, and resource usage
// @Tags			Vhosts
// @Produce		json
// @Param			vhost-name	path		string	true	"Vhost Name"
// @Success		200			{object}	models.VhostMetrics
// @Failure		400			{object}	httpsuite.ErrorResponse
// @Failure		401			{object}	httpsuite.ErrorResponse
// @Failure		403			{object}	httpsuite.ErrorResponse
// @Failure		502			{object}	httpsuite.ErrorResponse
// @Router			/v1/vhosts/{vhost-name}/metrics [get]
// @security		bearer
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
func (rc *RMQHandler) MetricHandler(w http.ResponseWriter, r *http.Request) {
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
	metrics, err := rc.RMQClient.GetMetrics(filter)
	if err != nil {
		httpsuite.WriteJSONError(w,
			http.StatusInternalServerError,
			httpsuite.WithError(err),
			httpsuite.WithErrorMessage("failed to fetch metrics for vhost"),
		)
		return
	}
	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &metrics)
}
