package api

import (
	"net/http"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

// @Summary		Health check
// @Description	Returns a simple health check response to indicate that the service is running
// @Tags			Health
// @Produce		json
// @Success		200	{string}	string	"healthy"
// @Router			/healthz [get]
func (rc *APIService) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	httpsuite.SendEmptyResponse(r.Context(), w, "healthy", http.StatusOK)
}

// @Summary		Readiness check
// @Description	Checks the readiness of the service by verifying connectivity to RabbitMQ, MongoDB, and Dex
// @Tags			Health
// @Produce		json
// @Success		200	{object}	models.HealthStatus	"ready"
// @Failure		502	{object}	models.HealthStatus	"not ready"
// @Router			/readyz [get]
func (rc *APIService) ReadyzHandler(rmq *rmq.RMQHandler, dex *dex.DexClient) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		status := models.NewHealthStatus()

		err := rmq.RMQClient.Ping()
		if err != nil {
			status.RabbitMQ = models.StatusUnhealthy
		} else {
			status.RabbitMQ = models.StatusHealthy
		}

		err = rc.DB.Ping(r.Context(), 5)
		if err != nil {
			status.Database = models.StatusUnhealthy
		} else {
			status.Database = models.StatusHealthy
		}

		err = dex.Ping(r.Context())
		if err != nil {
			status.Dex = models.StatusUnhealthy
		} else {
			status.Dex = models.StatusHealthy
		}

		if !status.IsHealthy() {
			httpsuite.SendResponse(r.Context(), w, "not ready", http.StatusServiceUnavailable, &status)
		}

		httpsuite.SendResponse(r.Context(), w, "ready", http.StatusOK, &status)
	}
}
