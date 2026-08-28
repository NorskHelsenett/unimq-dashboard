package api

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
)

// @Summary		Get all vhosts
// @Description	Get a list of all vhosts in the RabbitMQ cluster
// @Tags			Vhosts
// @Produce		json
// @Success		200	{array}		[]models.Vhost
// @Failure		502	{object}	httpsuite.APIError
// @Router			/v1/vhosts [get]
// @security		bearer
func (rc *APIService) VhostsHandler(w http.ResponseWriter, r *http.Request) {
	vhosts, err := rc.RMQClient.GetVhosts()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhosts", http.StatusInternalServerError)
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
// @Failure		400			{object}	httpsuite.APIError
// @Failure		502			{object}	httpsuite.APIError
// @Router			/v1/vhosts/{vhost-name} [get]
// @security		bearer
func (rc *APIService) VhostHandler(w http.ResponseWriter, r *http.Request) {
	vhostName := chi.URLParam(r, "vhost")
	if vhostName == "" {
		httpsuite.WriteJSONError(w, "vhost name is required", http.StatusBadRequest)
		return
	}

	eVhostName, err := url.QueryUnescape(vhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error decoding vhost name", http.StatusBadRequest)
		return
	}

	vhostData, err := rc.RMQClient.GetVhost(eVhostName)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching vhost data", http.StatusBadGateway)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "", http.StatusOK, &vhostData)
}
