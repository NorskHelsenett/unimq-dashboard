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

		// Check for any scheduled maintenance and send notifications if there are any new ones.
		checkMaintenanceSchedules(c.Ctx, c.DB, urls)

		// Evaluate each rule for the vhost and send notifications if needed.
		for _, rule := range vhost.Rules {
			c.checkRule(rule, vhost.Name, urls, metrics, queues)
		}
	}
}

var (
	ErrNotificationRuleDisabled      = fmt.Errorf("notification rule is disabled")
	ErrNotificationRuleInMaintenance = fmt.Errorf("notification rule is in maintenance mode")
	ErrNotificationRuleNoChange      = fmt.Errorf("notification rule status has not changed")
)

type evaluationResult struct {
	Triggered bool
	Value     *float64
	NewStatus models.AlarmStatus
}

// EvaluateMetrics evaluates a single alarm rule against the current metrics and returns whether it is triggered, the current value, and the new status.
func EvaluateMetrics(rule *models.AlarmRule, metrics *models.VhostMetrics, queues []models.QueueDetail) (*evaluationResult, error) {

	if !rule.Enabled {
		return nil, ErrNotificationRuleDisabled
	}
	if rule.Type == models.AlarmTypeMaintenance {
		return nil, ErrNotificationRuleInMaintenance
	}
	triggered, value := evaluate(rule, metrics, queues)
	newStatus := models.AlarmStatusOK
	if triggered {
		newStatus = models.AlarmStatusFiring
	}

	output := &evaluationResult{
		Triggered: triggered,
		Value:     value,
		NewStatus: newStatus,
	}

	return output, nil
}

// EvaluateRule checks if the status of the rule has changed and returns an AlarmEntry if it has.
func EvaluateRule(rule *models.AlarmRule, newStatus models.AlarmStatus, newValue float64) (*models.AlarmEntry, error) {

	if !rule.Enabled {
		return nil, ErrNotificationRuleDisabled
	}

	if rule.Status != models.AlarmStatusFiring && newStatus == models.AlarmStatusFiring {
		entry := models.NewLogEntry(models.LogEventFired, &newValue, rule.Threshold, rule.Type)
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}

		return &alarm, nil

	} else if rule.Status == models.AlarmStatusFiring && newStatus == models.AlarmStatusOK {
		entry := models.NewLogEntry(models.LogEventResolved, &newValue, rule.Threshold, rule.Type)
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}
		return &alarm, nil
	}

	return nil, ErrNotificationRuleNoChange
}

// checkRule evaluates a single alarm rule against the current metrics and sends notifications if needed.
func (c *Checker) checkRule(rule *models.AlarmRule, vhostName string, urls []string, metrics *models.VhostMetrics, queues []models.QueueDetail) {

	evalResult, err := EvaluateMetrics(rule, metrics, queues)
	if err != nil {
		switch err {
		case ErrNotificationRuleDisabled:
			slog.DebugContext(c.Ctx, "Skipping disabled rule", "vhost", vhostName, "rule", rule.Name)
			return
		case ErrNotificationRuleInMaintenance:
			slog.DebugContext(c.Ctx, "Skipping maintenance rule evaluation", "vhost", vhostName, "rule", rule.Name)
			return
		default:
			slog.ErrorContext(c.Ctx, "Failed to evaluate rule", "vhost", vhostName, "rule", rule.Name, "error", err)
			return
		}
	}

	shouldNotify := evalResult.Triggered && rule.Status != models.AlarmStatusFiring && len(urls) > 0
	slog.DebugContext(c.Ctx, "Evaluating rule",
		"vhost", vhostName,
		"rule", rule.Name,
		"type", rule.Type,
		"value", *evalResult.Value,
		"threshold", rule.Threshold,
		"triggered", evalResult.Triggered,
		// "urls", len(urls),
	)

	// If the status is changing to firing, it will update the LastFired timestamp.
	err = c.DB.UpdateNotificationRule(
		c.Ctx,
		vhostName,
		rule.ID,
		evalResult.NewStatus,
		*evalResult.Value,
		shouldNotify,
	)
	if err != nil {
		slog.ErrorContext(c.Ctx, "Failed to update notification rule status", "vhost", vhostName, "rule", rule.Name, "error", err)
	}

	alarm, err := EvaluateRule(rule, evalResult.NewStatus, *evalResult.Value)
	if err != nil {
		if err != ErrNotificationRuleNoChange {
			slog.ErrorContext(c.Ctx, "Failed to evaluate rule for alarm entry", "vhost", vhostName, "rule", rule.Name, "error", err)
		}
		return
	}
	err = c.DB.AddAlarm(c.Ctx, alarm)
	if err != nil {
		slog.ErrorContext(c.Ctx, "notify: failed to add alarm entry", "error", err)
	}

	if shouldNotify {
		err = Notify(c.Ctx, urls, rule, vhostName)
		if err != nil {
			slog.ErrorContext(c.Ctx, "notify: failed to send notification", "vhost", vhostName, "rule", rule.Name, "error", err)
		}
	}
}

// Notify sends a notification to the provided URLs with the alarm rule and vhost name.
func Notify(ctx context.Context, urls []string, rule *models.AlarmRule, vhostName string) error {
	subject := fmt.Sprintf("[UniMQ] Alarm: %s — %s", rule.Name, vhostName)
	body := rule.BuildMessage(vhostName)
	err := notificationhelper.SendWebhooks(urls, subject, body)
	if err != nil {
		return err
	}

	return nil
}

// nolint:gocyclo // While it is marked as complex, it only evaluates a single rule against the current metrics and returns whether it is triggered and the current value.
func evaluate(rule *models.AlarmRule, metrics *models.VhostMetrics, queues []models.QueueDetail) (bool, *float64) {
	var v float64
	switch rule.Type {
	case models.AlarmTypeChannels:
		if metrics != nil {
			v = float64(metrics.Channels)
		}
	case models.AlarmTypeConnections:
		if metrics != nil {
			v = float64(metrics.Connections)
		}
	case models.AlarmTypeQueues:
		if metrics != nil {
			v = float64(metrics.Queues)
		}
	case models.AlarmTypeUnacked:
		if metrics != nil {
			v = float64(metrics.Unacked)
		}
	case models.AlarmTypeQueueMessages:
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v = float64(q.Messages)
				break
			}
		}
	case models.AlarmTypeQueueSize:
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v = float64(q.MessageBytes)
				break
			}
		}
	case models.AlarmTypeNoConsumer:
		for _, q := range queues {
			if q.Name == rule.QueueName {
				v := float64(q.Messages)
				return q.Messages > 0 && q.Consumers == 0, &v
			}
		}
	default:
		slog.Error("Unknown rule type", "type", rule.Type)
		v := float64(0)
		return false, &v
	}

	return v >= rule.Threshold, &v
}

// checkMaintenanceRule checks for any scheduled maintenance and sends notifications if there are any new ones.
func checkMaintenanceSchedules(ctx context.Context, db *database.Database, urls []string) {
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
		if err := notificationhelper.SendWebhooks(urls, subject, body); err != nil {
			slog.ErrorContext(ctx, "notify: maintenance webhook failed", "error", err)
		} else {
			slog.InfoContext(ctx, "notify: maintenance webhook sent", "id", m.ID)
		}
		if err := db.SetMaintenanceEntryNotified(ctx, m.ID, true); err != nil {
			slog.ErrorContext(ctx, "Failed to mark maintenance as notified", "error", err)
		}
	}
}
