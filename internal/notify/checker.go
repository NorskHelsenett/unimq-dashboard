package notify

import (
	"context"
	"errors"
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
		Ctx         context.Context
		DB          *database.Database
		RMQClient   *rabbitmq.RMQClient
		interval    time.Duration
		mu          sync.RWMutex
		lastChecked time.Time
		runtimeMs   int64
		hasRun      bool
	}

	CheckerStatus struct {
		LastChecked *time.Time `json:"last_checked"`
		RuntimeMs   *int64     `json:"runtime_ms"`
		IntervalS   int64      `json:"interval_s"`
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

func (c *Checker) GetStatus() CheckerStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.hasRun {
		return CheckerStatus{IntervalS: int64(c.interval.Seconds())}
	}
	t := c.lastChecked
	ms := c.runtimeMs
	return CheckerStatus{LastChecked: &t, RuntimeMs: &ms, IntervalS: int64(c.interval.Seconds())}
}

func (c *Checker) StartChecker(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		// Initial delay to allow other components to start and populate the store before checks run.
		initTicker := time.NewTicker(15 * time.Second)
		defer initTicker.Stop()

		select {
		case <-initTicker.C:
			slog.InfoContext(c.Ctx, "Checker started")
		case <-c.Ctx.Done():
			slog.InfoContext(c.Ctx, "Checker stopped before first run")
			return
		}

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {

			select {
			case <-ticker.C:
				timer := time.Now()
				c.runChecks()
				elapsed := time.Since(timer)
				slog.InfoContext(c.Ctx, "finished checking maintenance statuses, notifications and metrics values", "runtime", elapsed)
				c.mu.Lock()
				c.lastChecked = time.Now()
				c.runtimeMs = elapsed.Milliseconds()
				c.hasRun = true
				c.mu.Unlock()
			case <-c.Ctx.Done():
				slog.InfoContext(c.Ctx, "Checker stopped")
				return
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

	if _, err := c.DB.AdvanceMaintenanceStatuses(c.Ctx); err != nil {
		slog.ErrorContext(c.Ctx, "Checker: failed to advance maintenance statuses", "error", err)
	}

	notifications, err := c.DB.GetNotificationsAll(c.Ctx)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to fetch notifications from database", "error", err)
		return
	}

	urls := make([]string, 0)
	emails := make([]string, 0)

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

		urls = append(urls, vhost.WebhookURLs()...)
		emails = append(emails, vhost.EmailRecipients()...)

		// Evaluate each rule for the vhost and send notifications if needed.
		for _, rule := range vhost.Rules {
			c.checkRule(rule, &vhost, metrics, queues)
		}
	}

	// Check for any scheduled maintenance and send notifications if there are any new ones.
	checkMaintenanceSchedules(c.Ctx, c.DB, urls, emails)
}

var (
	ErrNotificationRuleDisabled         = fmt.Errorf("notification rule is disabled")
	ErrNotificationRuleInMaintenance    = fmt.Errorf("notification rule is in maintenance mode")
	ErrNotificationRuleNoChange         = fmt.Errorf("notification rule status has not changed")
	ErrNotificationRuleUnknownType      = fmt.Errorf("notification rule has unknown type")
	ErrNotificationRuleQueueNotFound    = fmt.Errorf("queue not found")
	ErrNotificationRuleNoMetrics        = fmt.Errorf("no metrics available for evaluation")
	ErrNotificationRuleNoqueueMetrics   = fmt.Errorf("no queue metrics available for evaluation")
	ErrNotificationRuleEvaluationFailed = fmt.Errorf("notification rule evaluation failed")
)

// checkRule evaluates a single alarm rule against the current metrics and sends notifications if needed.
func (c *Checker) checkRule(rule *models.AlarmRule,
	vhost *models.VhostNotification,
	metrics *models.VhostMetrics,
	queues []models.QueueDetail,
) {

	evalResult, err := EvaluateMetrics(rule, metrics, queues)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotificationRuleDisabled):
			slog.DebugContext(c.Ctx, "Skipping disabled rule", "vhost", vhost.Name, "rule", rule.Name)
			return
		case errors.Is(err, ErrNotificationRuleInMaintenance):
			slog.DebugContext(c.Ctx, "Skipping maintenance rule evaluation", "vhost", vhost.Name, "rule", rule.Name)
			return
		case errors.Is(err, ErrNotificationRuleNoMetrics):
			slog.ErrorContext(c.Ctx, "Skipping rule evaluation due to missing metrics", "vhost", vhost.Name, "rule", rule.Name)
			errEntry := models.NewLogEntry(models.LogEventError, nil, rule.Threshold, rule.Type)
			err = c.DB.InsertAlarmEntries(c.Ctx, rule.ID, []models.LogEntry{errEntry})
			if err != nil {
				slog.ErrorContext(c.Ctx, "notify: failed to insert alarm entry", "error", err)
			}
			return
		case errors.Is(err, ErrNotificationRuleNoqueueMetrics):
			errEntry := models.NewLogEntry(models.LogEventError, nil, rule.Threshold, rule.Type)
			err = c.DB.InsertAlarmEntries(c.Ctx, rule.ID, []models.LogEntry{errEntry})
			if err != nil {
				slog.ErrorContext(c.Ctx, "notify: failed to insert alarm entry", "error", err)
			}
			slog.ErrorContext(c.Ctx, "Skipping rule evaluation due to missing queue metrics", "vhost", vhost.Name, "rule", rule.Name)
			return
		default:
			slog.ErrorContext(c.Ctx, "Failed to evaluate rule", "vhost", vhost.Name, "rule", rule.Name, "error", err)
			return
		}
	}

	shouldNotify := evalResult.Triggered && rule.Status != models.AlarmStatusFiring && len(vhost.WebhookURLs()) > 0
	slog.DebugContext(c.Ctx, "Evaluating rule",
		"vhost", vhost.Name,
		"rule", rule.Name,
		"type", rule.Type,
		"value", *evalResult.Value,
		"threshold", rule.Threshold,
		"triggered", evalResult.Triggered,
	)

	// If the status is changing to firing, it will update the LastFired timestamp.
	err = c.DB.UpdateNotificationRule(
		c.Ctx,
		vhost.Name,
		rule.ID,
		evalResult.NewStatus,
		*evalResult.Value,
		shouldNotify,
	)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to update notification rule status",
			"vhost", vhost.Name,
			"rule", rule.Name,
			"error", err,
		)
	}

	alarm, err := EvaluateRule(rule, evalResult.NewStatus, *evalResult.Value)
	if err != nil {
		if !errors.Is(err, ErrNotificationRuleNoChange) {
			slog.ErrorContext(c.Ctx, "Failed to evaluate rule for alarm entry",
				"vhost", vhost.Name,
				"rule", rule.Name,
				"error", err,
			)
		}
		return
	}

	err = c.DB.InsertAlarmEntries(c.Ctx, alarm.AlarmID, alarm.Entries)
	if err != nil {
		slog.ErrorContext(c.Ctx, "notify: failed to insert alarm entry", "error", err)
	}

	if shouldNotify {
		err = NotifyAlarm(c.Ctx, vhost, rule)
		if err != nil {
			slog.ErrorContext(c.Ctx, "notify: failed to send notification", "vhost", vhost.Name, "rule", rule.Name, "error", err)
		}
	}
}

// Notify sends a notification to the provided URLs and emails with the alarm rule and vhost name.
func NotifyAlarm(ctx context.Context, vhost *models.VhostNotification, rule *models.AlarmRule) error {
	subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhost.Name)
	body := rule.BuildMessage(vhost.Name)
	if len(vhost.WebhookURLs()) != 0 {
		err := notificationhelper.SendWebhooks(vhost.WebhookURLs(), subject, body)
		if err != nil {
			return err
		}
	}

	if len(vhost.EmailRecipients()) != 0 {
		for _, email := range vhost.EmailRecipients() {
			err := notificationhelper.EmailSenderInstance.SendEmail(
				email,
				subject,
				body,
				"text/plain",
			)
			if err != nil {
				if errors.Is(err, notificationhelper.ErrEmailNotConfigured) {
					slog.WarnContext(ctx, "notify: email not sent, SMTP server is not configured", "email", email)
					return err
				}
				slog.ErrorContext(ctx, "notify: failed to send email", "email", email, "error", err)
				continue
			}
		}
	}

	return nil
}

// checkMaintenanceRule checks for any scheduled maintenance and sends notifications if there are any new ones.
func checkMaintenanceSchedules(ctx context.Context, db *database.Database, urls []string, emails []string) {
	scheduled, err := db.GetMaintenanceScheduled(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch scheduled maintenance", "error", err)
		return
	}
	for i := range scheduled {
		m := &scheduled[i]
		if m.Notified {
			continue
		}

		body := fmt.Sprintf(
			"New maintenance scheduled:\n\n%s\n\nDate: %s – %s UTC",
			m.Description,
			m.Start.Format("2006-01-02 15:04"),
			m.End.Format("15:04"),
		)
		subject := "[UniMQ] New maintenance scheduled"
		err := notificationhelper.SendWebhooks(urls, subject, body)
		if err != nil {
			slog.ErrorContext(ctx, "notify: maintenance webhook failed", "error", err)
		} else {
			slog.InfoContext(ctx, "notify: maintenance webhook sent", "id", m.ID)
		}

		if err := db.SetMaintenanceEntryNotified(ctx, m.ID, true); err != nil {
			slog.ErrorContext(ctx, "Failed to mark maintenance as notified", "error", err)
		}

		err = notificationhelper.EmailSenderInstance.SendEmails(ctx, emails, subject, body, "text/plain")
		if err != nil {
			if errors.Is(err, notificationhelper.ErrEmailNotConfigured) {
				slog.WarnContext(ctx, "maintenance email not sent, SMTP server is not configured", "emails", emails)
				return
			}
			slog.ErrorContext(ctx, "maintenance email failed on some", "error", err)

		}

		slog.InfoContext(ctx, "notify: maintenance email sent", "emails", emails)
	}
}
