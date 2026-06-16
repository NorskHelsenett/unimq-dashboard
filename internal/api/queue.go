package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

// @Summary Get Queue Metrics for a specific queue in the specified virtual host
// @Description Fetches time-series metrics for a specific RabbitMQ queue over a specified time range.
// @Tags Metrics
// @Produce html
// @Param vhost query string true "Virtual Host"
// @Param name query string true "Queue Name"
// @Param range query string false "Time Range (e.g., 1h, 24h, 7d)" default(1h)
// @Success 200 {string} string "HTML page with queue metrics"
// @Failure 400 {object} httpsuite.ErrorResponse "Bad Request"
// @Failure 500 {object} httpsuite.ErrorResponse "Internal Server Error"
// @Router /queue [get]
func (rc *APIService) QueueHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	queue := r.URL.Query().Get("name")
	if queue == "" {
		httpsuite.WriteJSONError(w, "missing queue name", http.StatusBadRequest)
		return
	}

	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "1h"
	}

	since, ok := models.RangeDurations[rangeStr]
	if !ok {
		since = time.Hour
		rangeStr = "1h"
	}

	samples, err := rc.PromClient.QueryRange(models.RangeOptions{
		Vhost: vhost,
		Queue: queue,
		Since: since,
		Step:  prometheus.StepFor(since),
	})

	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queue metrics", http.StatusInternalServerError)
		return
	}
	samplesJSON := "[]"
	if len(samples) > 0 {
		b, _ := json.Marshal(samples)
		samplesJSON = string(b)
	}
	data := models.NewQueueData(vhost, queue, rangeStr, samplesJSON, samples)
	if err := templating.QueueTmpl.Execute(w, data); err != nil {
		httpsuite.WriteJSONError(w, "error rendering template", http.StatusInternalServerError)
		return
	}
}

// @Summary Get All Queues to a specified virtual host
// @Description Fetches details of all queues in a specified virtual host.
// @Tags Queues
// @Produce json
// @Param vhost query string true "Virtual Host"
// @Success 200 {array} models.QueueDetail "List of queue details"
// @Failure 400 {object} httpsuite.ErrorResponse "Bad Request"
// @Failure 500 {object} httpsuite.ErrorResponse "Internal Server Error"
// @Router /queues [get]
func (rc *APIService) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}

	details, err := rc.RMQClient.GetQueues()
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queue details", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queue details", http.StatusOK, &details)
}
