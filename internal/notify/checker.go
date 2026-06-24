package notify

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/notificationhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
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

func WithContext(ctx context.Context) CheckerOptions {
	return func(c *Checker) {
		c.Ctx = ctx
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

func (c *Checker) StartChecker(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		// Initial delay to allow other components to start and populate the store before checks run.
		initTicker := time.NewTicker(15 * time.Second)
		select {
		case <-initTicker.C:
			slog.InfoContext(c.Ctx, "Checker started")
		case <-c.Ctx.Done():
			slog.InfoContext(c.Ctx, "Checker stopped before first run")
			return
		}

		ticker := time.NewTicker(c.interval)
		for {

			select {
			case <-ticker.C:
				slog.DebugContext(c.Ctx, "Checker tick")
			case <-c.Ctx.Done():
				slog.InfoContext(c.Ctx, "Checker stopped")
				return
			}

			select {
			case <-c.Ctx.Done():
				slog.InfoContext(c.Ctx, "Checker stopped")
				return
			default:
				c.runChecks()
			}
		}
	}()
}

// runChecks fetches notifications and metrics, evaluates rules, updates statuses, and sends notifications as needed.
// We don't return an error as we don't care about individual failures here - we just want to log them and keep going.
func (c *Checker) runChecks() {
	if !c.DB.Inialized {
		slog.WarnContext(c.Ctx, "Checker: database not initialized")
		return
	}
	notifications, err := c.DB.GetNotificationsAll(c.Ctx)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to fetch notifications from database", "error", err)
		return
	}

	for _, vhost := range notifications {
		if len(vhost.Rules) == 0 {
			continue
		}

		metrics, err := c.RMQClient.GetMetrics(vhost.Name)
		if err != nil {
			slog.ErrorContext(c.Ctx, "Failed to fetch metrics", "vhost", vhost.Name, "error", err)
			continue
		}

		queues, err := c.RMQClient.GetQueueDetails(vhost.Name)
		if err != nil {
			slog.ErrorContext(c.Ctx, "Failed to fetch queue details", "vhost", vhost.Name, "error", err)
			continue
		}

		urls := vhost.WebhookURLs()
		for _, rule := range vhost.Rules {
			c.checkRule(rule, vhost.Name, urls, metrics, queues)
		}
	}
}

func (c *Checker) checkRule(rule *models.AlarmRule, vhostName string, urls []string, metrics *models.VhostMetrics, queues []models.QueueDetail) {

	if !rule.Enabled {
		return
	}
	if rule.Type == models.AlarmTypeMaintenance {
		checkMaintenanceRule(c.Ctx, c.DB, vhostName, rule, urls)
		return
	}
	triggered, value := evaluate(*rule, metrics, queues)
	newStatus := models.AlarmStatusOK
	if triggered {
		newStatus = models.AlarmStatusFired
	}
	shouldNotify := triggered && rule.Status != models.AlarmStatusFiring && len(urls) > 0
	slog.DebugContext(c.Ctx, "Evaluating rule",
		"vhost", vhostName,
		"rule", rule.Name,
		"type", rule.Type,
		"value", *value,
		"threshold", rule.Threshold,
		"triggered", triggered,
		"urls", len(urls),
	)

	// TODO: This shouldn't run on every check, only on status changes.
	// This would require a local cache of the last known status, or a more complex DB query to check if the status has changed.
	err := c.DB.UpdateNotificationRule(
		c.Ctx,
		vhostName,
		rule.ID,
		newStatus,
		*value,
		shouldNotify,
	)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to update notification rule status", "vhost", vhostName, "rule", rule.Name, "error", err)
	}

	if rule.Status != models.AlarmStatusFiring && newStatus == models.AlarmStatusFiring {
		entry := models.LogEntry{Timestamp: time.Now(), Event: models.LogEventFired, Value: value, Threshold: rule.Threshold}
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}
		err := c.DB.AddAlarm(c.Ctx, &alarm)
		if err != nil {
			slog.ErrorContext(c.Ctx, "notify: failed to add alarm entry", "error", err)
		}

	} else if rule.Status == models.AlarmStatusFiring && newStatus == models.AlarmStatusOK {
		entry := models.LogEntry{Timestamp: time.Now(), Event: models.LogEventResolved, Value: value, Threshold: rule.Threshold}
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}
		err = c.DB.AddAlarm(c.Ctx, &alarm)
		if err != nil {
			slog.ErrorContext(c.Ctx, "notify: failed to add alarm entry", "error", err)
		}
	}

	if shouldNotify {
		subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhostName)
		body := rule.BuildMessage(vhostName)
		err := notificationhelper.SendWebhooks(urls, subject, body)
		if err != nil {
			slog.ErrorContext(c.Ctx, "notify: webhook failed", "rule", rule.Name, "error", err)
		} else {
			slog.InfoContext(c.Ctx, "notify: webhook sent", "rule", rule.Name, "vhost", vhostName)
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
	default:
		slog.Error("Unknown rule type", "type", rule.Type)
		v := float64(0)
		return false, val(v)
	}

	return false, nil
}

func checkMaintenanceRule(ctx context.Context, db *database.Database, vhost string, rule *models.AlarmRule, urls []string) {
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
				"New maintenance scheduled:\n\n%s\n\nDate: %s – %s UTC",
				m.Description,
				m.Start.Format("2006-01-02 15:04"),
				m.End.Format("15:04"),
			)
		}
		subject := "[UniMQ] New maintenance scheduled"
		if err := notificationhelper.SendWebhooks(urls, subject, body); err != nil {
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

	status := models.MaintenanceStatusDone
	if fired {
		status = models.MaintenanceStatusScheduled
	}

	err = db.SetMaintenanceEntryStatus(ctx, vhost, status)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update maintenance status", "error", err)
	}
}
