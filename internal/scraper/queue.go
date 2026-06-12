package scraper

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) QueueHandler(w http.ResponseWriter, r *http.Request) {
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

	samples, err := rc.PromClient.QueryRange(prom.RangeOptions{
		Vhost: vhost,
		Queue: queue,
		Since: since,
		Step:  prom.StepFor(since),
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

func (rc *RestClient) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		httpsuite.WriteJSONError(w, "missing vhost", http.StatusBadRequest)
		return
	}
	details, err := rc.GetQueueDetails(vhost)
	if err != nil {
		httpsuite.WriteJSONError(w, "error fetching queue details", http.StatusInternalServerError)
		return
	}

	httpsuite.SendResponse(r.Context(), w, "fetched queue details", http.StatusOK, &details)
}
