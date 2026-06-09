package notify

import (
	"fmt"
	"log"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/maintenance"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify/store"
	"github.com/sisneve/rabbitmq-dashboard/internal/scraper"
)

type (
	Checker struct {
		Config     *config.Config
		store      *store.Store
		logStore   *store.LogStore
		maintStore *maintenance.Store
		interval   time.Duration
	}

	CheckerOptions func(*Checker)
)

func WithConfig(cfg *config.Config) CheckerOptions {
	return func(c *Checker) {
		c.Config = cfg
	}
}

func WithStore(s *store.Store) CheckerOptions {
	return func(c *Checker) {
		c.store = s
	}
}

func WithLogStore(ls *store.LogStore) CheckerOptions {
	return func(c *Checker) {
		c.logStore = ls
	}
}

func WithMaintStore(ms *maintenance.Store) CheckerOptions {
	return func(c *Checker) {
		c.maintStore = ms
	}
}

func WithInterval(d time.Duration) CheckerOptions {
	return func(c *Checker) {
		c.interval = d
	}
}

func NewChecker(opts ...CheckerOptions) *Checker {
	c := &Checker{
		Config:     config.NewConfig(),
		store:      nil,
		logStore:   nil,
		maintStore: nil,
		interval:   60 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Checker) StartChecker() {
	go func() {
		time.Sleep(15 * time.Second)
		for {
			c.runChecks()
			time.Sleep(c.interval)
		}
	}()
}

func (c *Checker) runChecks() {
	snapshots := c.store.AllSnapshots()
	for vhost, snap := range snapshots {
		if len(snap.Rules) == 0 {
			continue
		}
		urls := snap.WebhookURLs()
		var metrics *models.VhostMetrics
		var queues []models.QueueDetail
		fetched := false
		fetchOnce := func() {
			if fetched {
				return
			}
			restclient := scraper.NewRestClient(
				fmt.Sprintf("%v:%d/api",
					c.Config.RabbitMQURL,
					c.Config.RabbitMQPort,
				),
				c.Config.RabbitMQUsername,
				c.Config.RabbitMQPassword,
				c.Config.PrometheusURL,
				"v1",
				c.Config.PrometheusPort,
			)

			metrics, _ = restclient.GetMetrics(vhost)
			queues, _ = restclient.GetQueueDetails(vhost)
			fetched = true
		}
		for _, rule := range snap.Rules {
			if !rule.Enabled {
				continue
			}
			if rule.Type == "maintenance" {
				checkMaintenanceRule(c.store, vhost, rule, urls, c.maintStore)
				continue
			}
			fetchOnce()
			triggered, value := evaluate(rule, metrics, queues)
			newStatus := "ok"
			if triggered {
				newStatus = "firing"
			}
			shouldNotify := triggered && rule.Status != "firing" && len(urls) > 0
			if err := c.store.SetRuleStatus(vhost, rule.ID, newStatus, shouldNotify, value); err != nil {
				log.Printf("notify: set status failed: %v", err)
			}
			// Log state transitions
			if rule.Status != "firing" && newStatus == "firing" {
				entry := store.LogEntry{Timestamp: time.Now(), Event: store.LogEventFired, Value: value, Threshold: rule.Threshold}
				if err := c.logStore.Append(rule.ID, entry); err != nil {
					log.Printf("notify: log append failed: %v", err)
				}
			} else if rule.Status == "firing" && newStatus == "ok" {
				entry := store.LogEntry{Timestamp: time.Now(), Event: store.LogEventResolved, Value: value, Threshold: rule.Threshold}
				if err := c.logStore.Append(rule.ID, entry); err != nil {
					log.Printf("notify: log append failed: %v", err)
				}
			}
			if shouldNotify {
				subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhost)
				body := rule.BuildMessage(vhost)
				if err := c.store.SendWebhooks(urls, subject, body); err != nil {
					log.Printf("notify: webhook failed (rule %s): %v", rule.Name, err)
				} else {
					log.Printf("notify: webhook sent for rule %q on vhost %q", rule.Name, vhost)
				}
			}
		}
	}
}

func evaluate(rule models.AlarmRule, metrics *models.VhostMetrics, queues []models.QueueDetail) (bool, *float64) {
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

func checkMaintenanceRule(store *store.Store, vhost string, rule models.AlarmRule, urls []string, maintStore *maintenance.Store) {
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
