package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/store/notify"
)

type (
	// Checker periodically evaluates alarm rules against current RabbitMQ metrics and triggers notifications.
	Checker struct {
		Ctx       context.Context
		DB        *database.Database
		RMQClient *rabbitmq.RMQClient
		interval  time.Duration
	}

	CheckerOptions func(*Checker)
)

func WithRMQClient(client *rabbitmq.RMQClient) CheckerOptions {
	return func(c *Checker) {
		c.RMQClient = client
	}
}

func WithDB(db *database.Database) CheckerOptions {
	return func(c *Checker) {
		c.DB = db
	}
}

func WithInterval(d time.Duration) CheckerOptions {
	return func(c *Checker) {
		c.interval = d
	}
}

func NewChecker(opts ...CheckerOptions) *Checker {
	c := &Checker{
		interval: 60 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Checker) StartChecker() {
	go func() {
		// Initial delay to allow other components to start and populate the store before checks run.
		time.Sleep(15 * time.Second)
		for {
			c.runChecks()
			time.Sleep(c.interval)
		}
	}()
}

func (c *Checker) runChecks() {
	notifications, err := c.DB.GetNotificationsAll(c.Ctx)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to fetch notifications from database", "error", err)
		return
	}

	for _, vhost := range notifications {
		if len(vhost.Rules) == 0 {
			continue
		}
		urls := vhost.WebhookURLs()

		metrics, err := c.RMQClient.GetMetrics(vhost.Name)
		if err != nil {
			slog.ErrorContext(c.Ctx, "Failed to fetch metrics", "vhost", vhost, "error", err)
			continue
		}

		queues, err := c.RMQClient.GetQueueDetails(vhost.Name)
		if err != nil {
			slog.ErrorContext(c.Ctx, "Failed to fetch queue details", "vhost", vhost, "error", err)
			continue
		}

		for _, rule := range vhost.Rules {

			if !rule.Enabled {
				continue
			}
			if rule.Type == "maintenance" {
				checkMaintenanceRule(c.Ctx, c.DB, vhost.Name, rule, urls)
				continue
			}
			triggered, value := evaluate(rule, metrics, queues)
			newStatus := "ok"
			if triggered {
				newStatus = "firing"
			}
			shouldNotify := triggered && rule.Status != "firing" && len(urls) > 0
			err := c.DB.UpdateNotificationRule(c.Ctx, vhost.Name, rule.ID, newStatus, *value, shouldNotify)
			if err != nil {
				log.Printf("notify: database update failed: %v", err)
			}

			if rule.Status != "firing" && newStatus == "firing" {
				entry := notify.LogEntry{Timestamp: time.Now(), Event: notify.LogEventFired, Value: value, Threshold: rule.Threshold}
				alarm := database.AlarmEntry{
					AlarmID: rule.ID,
					Entries: []notify.LogEntry{entry},
				}
				err := c.DB.AddAlarm(c.Ctx, &alarm)
				if err != nil {
					slog.ErrorContext(c.Ctx, "notify: failed to add alarm entry", "error", err)
				}

			} else if rule.Status == "firing" && newStatus == "ok" {
				entry := notify.LogEntry{Timestamp: time.Now(), Event: notify.LogEventResolved, Value: value, Threshold: rule.Threshold}
				alarm := database.AlarmEntry{
					AlarmID: rule.ID,
					Entries: []notify.LogEntry{entry},
				}
				err = c.DB.AddAlarm(c.Ctx, &alarm)
				if err != nil {
					slog.ErrorContext(c.Ctx, "notify: failed to add alarm entry", "error", err)
				}
			}

			if shouldNotify {
				subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhost.Name)
				body := rule.BuildMessage(vhost.Name)
				err := sendWebhooks(urls, subject, body)
				if err != nil {
					slog.ErrorContext(c.Ctx, "notify: webhook failed", "rule", rule.Name, "error", err)
				} else {
					slog.InfoContext(c.Ctx, "notify: webhook sent", "rule", rule.Name, "vhost", vhost)
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

func checkMaintenanceRule(ctx context.Context, db *database.Database, vhost string, rule models.AlarmRule, urls []string) {
	scheduled, err := db.GetMaintenanceScheduled(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch scheduled maintenance", "error", err)
		return
	}
	fired := false
	for _, m := range scheduled {
		if m.Notified {
			continue
		}

		body := rule.Message
		if body == "" {
			body = fmt.Sprintf(
				"New maintenace scheduled:\n\n%s\n\nDate: %s – %s UTC",
				m.Description,
				m.Start.Format("2006-01-02 15:04"),
				m.End.Format("15:04"),
			)
		}
		subject := "[UniMQ] New maintenance scheduled"
		if err := sendWebhooks(urls, subject, body); err != nil {
			slog.ErrorContext(ctx, "notify: maintenance webhook failed", "error", err)
		} else {
			slog.InfoContext(ctx, "notify: maintenance webhook sent", "id", m.ID)
		}
		err = db.SetMaintenanceEntryNotified(ctx, m.ID, true)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to mark maintenance as notified", "error", err)
		}
		fired = true
	}

	status := "ok"
	if fired {
		status = "firing"
	}

	err = db.SetMaintenanceEntryStatus(ctx, vhost, status)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update maintenance status", "error", err)
	}
}

func sendWebhooks(urls []string, subject, body string) error {
	text := subject + "\n\n" + body
	payload, _ := json.Marshal(map[string]string{"text": text})
	var lastErr error
	for _, u := range urls {
		resp, err := http.Post(u, "application/json", bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("webhook returnerte HTTP %d", resp.StatusCode)
		}
	}
	return lastErr
}
