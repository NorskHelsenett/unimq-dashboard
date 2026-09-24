package rmq

import (
	"net/http"

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
// @security		OAuth2[openid, profile, email, groups, audience:server:client_id:unimq-dashboard]
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
