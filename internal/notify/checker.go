package notify

import (
	"fmt"
	"log"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/maintenance"
	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

func StartChecker(store *Store, maintStore *maintenance.Store, interval time.Duration) {
	go func() {
		time.Sleep(15 * time.Second)
		for {
			runChecks(store, maintStore)
			time.Sleep(interval)
		}
	}()
}

func runChecks(store *Store, maintStore *maintenance.Store) {
	snapshots := store.AllSnapshots()
	for vhost, snap := range snapshots {
		if len(snap.Rules) == 0 {
			continue
		}
		urls := snap.WebhookURLs()
		var metrics *scraper.VhostMetrics
		var queues []scraper.QueueDetail
		fetched := false
		fetchOnce := func() {
			if fetched {
				return
			}
			metrics, _ = scraper.GetMetrics(vhost)
			queues, _ = scraper.GetQueueDetails(vhost)
			fetched = true
		}
		for _, rule := range snap.Rules {
			if !rule.Enabled {
				continue
			}
			if rule.Type == "maintenance" {
				checkMaintenanceRule(store, vhost, rule, urls, maintStore)
				continue
			}
			fetchOnce()
			triggered, value := evaluate(rule, metrics, queues)
			newStatus := "ok"
			if triggered {
				newStatus = "firing"
			}
			shouldNotify := triggered && rule.Status != "firing" && len(urls) > 0
			if err := store.SetRuleStatus(vhost, rule.ID, newStatus, shouldNotify, value); err != nil {
				log.Printf("notify: set status failed: %v", err)
			}
			if shouldNotify {
				subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhost)
				body := BuildMessage(rule, vhost)
				if err := store.SendWebhooks(urls, subject, body); err != nil {
					log.Printf("notify: webhook failed (rule %s): %v", rule.Name, err)
				} else {
					log.Printf("notify: webhook sent for rule %q on vhost %q", rule.Name, vhost)
				}
			}
		}
	}
}

func evaluate(rule AlarmRule, metrics *scraper.VhostMetrics, queues []scraper.QueueDetail) (bool, *float64) {
	val := func(v float64) *float64 { return &v }
	switch rule.Type {
	case "channels":
		if metrics != nil {
			v := float64(metrics.Channels)
			return v >= rule.Threshold, val(v)
		}
	case "connections":
		if metrics != nil {
			v := float64(metrics.Connections)
			return v >= rule.Threshold, val(v)
		}
	case "queues":
		if metrics != nil {
			v := float64(metrics.Queues)
			return v >= rule.Threshold, val(v)
		}
	case "unacked":
		if metrics != nil {
			v := float64(metrics.Unacked)
			return v >= rule.Threshold, val(v)
		}
	case "queue_messages":
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v := float64(q.Messages)
				return v >= rule.Threshold, val(v)
			}
		}
	case "queue_size":
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v := float64(q.MessageBytes)
				return v >= rule.Threshold, val(v)
			}
		}
	case "no_consumer":
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v := float64(q.Messages)
				return q.Messages > 0 && q.Consumers == 0, val(v)
			}
		}
	}
	return false, nil
}

func checkMaintenanceRule(store *Store, vhost string, rule AlarmRule, urls []string, maintStore *maintenance.Store) {
	scheduled := maintStore.Scheduled()
	fired := false
	for _, m := range scheduled {
		if store.IsMaintenanceNotified(m.ID) {
			continue
		}
		body := rule.Message
		if body == "" {
			body = fmt.Sprintf(
				"Nytt vedlikehold er planlagt:\n\n%s\n\nTidspunkt: %s – %s UTC",
				m.Description,
				m.Start.Format("2006-01-02 15:04"),
				m.End.Format("15:04"),
			)
		}
		subject := "[UniMQ] Nytt vedlikehold planlagt"
		if err := store.SendWebhooks(urls, subject, body); err != nil {
			log.Printf("notify: maintenance webhook failed: %v", err)
		} else {
			log.Printf("notify: maintenance webhook sent for entry %s", m.ID)
		}
		store.MarkMaintenanceNotified(m.ID)
		fired = true
	}
	status := "ok"
	if fired {
		status = "firing"
	}
	store.SetRuleStatus(vhost, rule.ID, status, fired, nil)
}
