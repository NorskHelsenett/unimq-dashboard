package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
)

// @Summary		Health check
// @Description	Returns a simple health check response to indicate that the service is running
// @Tags			Health
// @Produce		json
// @Success		200	{string}	string	"healthy"
// @Router			/api/healthz [get]
func (rc *APIService) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	httpsuite.SendEmptyResponse(r.Context(), w, "healthy", http.StatusOK)
}

// @Summary		Readiness check
// @Description	Checks the readiness of the service by verifying connectivity to RabbitMQ, MongoDB, and Dex
// @Tags			Health
// @Produce		json
// @Success		200	{string}	string	"ready"
// @Failure		502	{object}	httpsuite.ErrorResponse
// @Router			/api/readyz [get]
func (rc *APIService) ReadyzHandler(dex *dex.DexClient) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := rc.RMQClient.Ping()
		if err != nil {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to ping RabbitMQ"),
			)
			return
		}

		err = rc.DB.Ping(r.Context(), 5)
		if err != nil {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to ping MongoDB"),
			)
			return
		}

		err = dex.Ping(r.Context())
		if err != nil {
			httpsuite.WriteJSONError(w,
				http.StatusInternalServerError,
				httpsuite.WithError(err),
				httpsuite.WithErrorMessage("failed to ping Dex"),
			)
			return
		}

		httpsuite.SendEmptyResponse(r.Context(), w, "ready", http.StatusOK)
	}
}
