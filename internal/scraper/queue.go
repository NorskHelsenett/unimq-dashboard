package scraper

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
	"github.com/sisneve/rabbitmq-dashboard/internal/templating"
)

func (rc *RestClient) QueueHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	queue := r.URL.Query().Get("name")
	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "1h"
	}
	since, ok := models.RangeDurations[rangeStr]
	if !ok {
		since = time.Hour
		rangeStr = "1h"
	}
	samples, err := rc.PromClient.QueryRange(prom.RangeOptions{Vhost: vhost, Queue: queue, Since: since, Step: prom.StepFor(since)})
	if err != nil {
		log.Printf("prometheus error: %v", err)
	}
	samplesJSON := "[]"
	if len(samples) > 0 {
		b, _ := json.Marshal(samples)
		samplesJSON = string(b)
	}
	data := struct {
		Vhost, Queue, SelectedRange string
		Ranges                      []models.RangeOption
		SamplesJSON                 template.JS
		NoData                      bool
	}{vhost, queue, rangeStr, models.TimeRanges, template.JS(samplesJSON), len(samples) == 0}
	if err := templating.QueueTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (rc *RestClient) GetQueuesHandler(w http.ResponseWriter, r *http.Request) {
	vhost := r.URL.Query().Get("vhost")
	if vhost == "" {
		http.Error(w, "missing vhost", http.StatusBadRequest)
		return
	}
	details, err := rc.GetQueueDetails(vhost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}
