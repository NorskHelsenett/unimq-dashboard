package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (dbc *Database) GetNotificationRules(ctx context.Context, notificationID string) ([]*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved notification rules", "runtime", time.Since(start), id, notificationID, "count", len(notification.Rules))
	return notification.Rules, nil
}

var (
	ErrNotificationRuleNotFound = fmt.Errorf("notification rule not found")
)

func (dbc *Database) GetNotificationRule(ctx context.Context, notificationID string, ruleID string) (*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, notificationID)
	if err != nil {
		if errors.Is(err, ErrVhostNotFound) {
			return nil, err
		}
		return nil, err
	}

	for _, rule := range notification.Rules {
		if rule.ID == ruleID {
			slog.DebugContext(ctx, "retrieved notification rule",
				"runtime", time.Since(start),
				"vhost", notificationID,
				"rule", rule.Name,
				id, rule.ID,
			)
			return rule, nil
		}
	}
	return nil, ErrNotificationRuleNotFound
}

func (dbc *Database) AddNotificationRule(ctx context.Context, vhost string, rule *models.AlarmRule) error {
	start := time.Now()

	filter := map[string]any{id: vhost}
	update := map[string]any{
		"$push": map[string]any{
			"rules": rule,
		},
	}

	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			"rule", rule.Name,
			"id", rule.ID,
			"error", err,
		)
		return err
	}

	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification found to add rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			"rule", rule.Name,
			"id", rule.ID,
		)
		return fmt.Errorf("notification not found for vhost %s. %w", vhost, ErrVhostNotFound)
	}

	slog.DebugContext(ctx, "added notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		"rule", rule.Name,
	)

	return err
}

func (dbc *Database) DeleteNotificationRule(ctx context.Context, vhost string, ruleID string) error {
	start := time.Now()

	filter := map[string]any{id: vhost}
	update := map[string]any{
		"$pull": map[string]any{
			"rules": map[string]any{"id": ruleID},
		},
	}

	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			id, ruleID,
			"error", err,
		)
		return fmt.Errorf("failed to delete notification rule. %w", err)
	}

	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification found to delete rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			id, ruleID,
		)
		return fmt.Errorf("notification not found for vhost %s. %w", vhost, ErrVhostNotFound)
	}

	slog.DebugContext(ctx, "deleted notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		id, ruleID,
	)
	return nil
}

func (dbc *Database) UpdateNotificationRule(ctx context.Context, vhost, ruleID string, status models.AlarmStatus, value float64, notified bool) error {
	start := time.Now()

	setFields := map[string]any{
		"rules.$.status":    status,
		"rules.$.lastValue": value,
		"notified":          notified,
	}
	if status == models.AlarmStatusFiring {
		setFields["rules.$.lastFired"] = time.Now()
	}

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{set: setFields}

	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule status",
			"runtime", time.Since(start),
			id, vhost,
			"rule", ruleID,
			"error", err,
		)
		return err
	}
	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification rule found to update",
			"runtime", time.Since(start),
			id, vhost,
			"rule", ruleID,
		)
		return fmt.Errorf("notification rule not found for vhost %s and rule %s. %w", vhost, ruleID, ErrNotificationRuleNotFound)
	}

	slog.DebugContext(ctx, "updated notification rule status",
		"runtime", time.Since(start),
		id, vhost,
		"ruleID", ruleID,
		"notified", notified,
	)

	return err
}

// TODO: Should just be a wrapper function for UpdateNotificationRule, but with a different name for clarity. Consider refactoring.
func (dbc *Database) ToggleNotificationRule(ctx context.Context, vhost, ruleID string, enabled bool) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.enabled": enabled,
		},
	}
	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to toggle notification rule",
			"runtime", time.Since(start),
			id, vhost,
			"ruleID", ruleID,
			"enabled", enabled,
			"error", err,
		)
		return err
	}

	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification rule found to toggle",
			"runtime", time.Since(start),
			id, vhost,
			"ruleID", ruleID,
			"enabled", enabled,
		)
		return fmt.Errorf("notification rule not found for vhost %s and rule %s. %w", vhost, ruleID, ErrNotificationRuleNotFound)
	}

	slog.DebugContext(ctx, "toggled notification rule",
		"runtime", time.Since(start),
		id, vhost, "ruleID",
		ruleID, "enabled", enabled,
	)
	return nil
}

// TODO: Consider refactoring into a put function.
func (dbc *Database) UpdateNotificationRuleThreshold(ctx context.Context, vhost, ruleID string, threshold float64) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.threshold": threshold,
		},
	}
	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule",
			"runtime", time.Since(start),
			"vhost", vhost, "ruleID",
			ruleID, "error", err,
		)
	}

	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification rule found to update",
			"runtime", time.Since(start),
			"vhost", vhost,
			"ruleID", ruleID,
		)
		return fmt.Errorf("notification rule not found for vhost %s and rule %s. %w", vhost, ruleID, ErrNotificationRuleNotFound)
	}
	slog.DebugContext(ctx, "updated notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		"ruleID", ruleID,
	)

	return err
}

// TODO: Consider refactoring into a put function.
func (dbc *Database) UpdateNotificationRuleMessage(ctx context.Context, vhost, ruleID string, message string) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.message": message,
		},
	}

	result, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule message",
			"runtime", time.Since(start),
			id, vhost,
			"ruleID", ruleID,
			"error", err,
		)
	}

	if result.ModifiedCount == 0 {
		slog.ErrorContext(ctx, "no notification rule found to update message",
			"runtime", time.Since(start),
			id, vhost,
			"ruleID", ruleID,
		)
		return fmt.Errorf("notification rule not found for vhost %s and rule %s. %w", vhost, ruleID, ErrNotificationRuleNotFound)
	}

	slog.DebugContext(ctx, "updated notification rule message",
		"runtime", time.Since(start),
		id, vhost,
		"ruleID", ruleID,
	)

	return err
}
