package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (dbc *Database) GetNotificationRules(ctx context.Context, id string) ([]*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, id)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved notification rules", "runtime", time.Since(start), id, id, "count", len(notification.Rules))
	return notification.Rules, nil
}

var (
	ErrNotificationRuleNotFound = fmt.Errorf("notification rule not found")
)

func (dbc *Database) GetNotificationRule(ctx context.Context, vhost string, ruleid string) (*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, vhost)
	if err != nil {
		if errors.Is(err, ErrVhostNotFound) {
			return nil, err
		}
		return nil, err
	}

	for _, rule := range notification.Rules {
		if rule.ID == ruleid {
			slog.DebugContext(ctx, "retrieved notification rule",
				"runtime", time.Since(start),
				"vhost", vhost,
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

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			"rule", rule.Name,
			"id", rule.ID,
			"error", err,
		)
	} else {
		slog.DebugContext(ctx, "added notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			"rule", rule.Name,
		)
	}

	return err
}

func (dbc *Database) DeleteNotificationRule(ctx context.Context, vhost string, id string) error {
	start := time.Now()

	filter := map[string]any{id: vhost}
	update := map[string]any{
		"$pull": map[string]any{
			"rules": map[string]any{"id": id},
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			id, id,
			"error", err,
		)
		return fmt.Errorf("failed to delete notification rule. %w", err)
	}

	slog.DebugContext(ctx, "deleted notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		id, id,
	)
	return nil
}

func (dbc *Database) UpdateNotificationRule(ctx context.Context, vhost, name string, status models.AlarmStatus, value float64, notified bool) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.name": name}
	update := map[string]any{
		set: map[string]any{
			"rules.$.status":    status,
			"rules.$.lastValue": value,
			"notified":          notified,
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule status", "runtime", time.Since(start), id, vhost, "rule", name, "error", err)
		return err
	} else {
		slog.DebugContext(ctx, "updated notification rule status", "runtime", time.Since(start), id, vhost, "name", name, "notified", notified)
	}

	return err
}

func (dbc *Database) ToggleNotificationRule(ctx context.Context, vhost, ruleID string, enabled bool) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.enabled": enabled,
		},
	}
	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to toggle notification rule", "runtime", time.Since(start), id, vhost, "ruleID", ruleID, "enabled", enabled, "error", err)
		return err
	}

	slog.DebugContext(ctx, "toggled notification rule", "runtime", time.Since(start), id, vhost, "ruleID", ruleID, "enabled", enabled)
	return err
}

func (dbc *Database) UpdateNotificationRuleThreshold(ctx context.Context, vhost, ruleID string, threshold float64) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.threshold": threshold,
		},
	}
	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID, "error", err)
	} else {
		slog.DebugContext(ctx, "updated notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID)
	}

	return err
}

func (dbc *Database) UpdateNotificationRuleMessage(ctx context.Context, vhost, ruleID string, message string) error {
	start := time.Now()

	filter := map[string]any{id: vhost, "rules.id": ruleID}
	update := map[string]any{
		set: map[string]any{
			"rules.$.message": message,
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule message", "runtime", time.Since(start), id, vhost, "ruleID", ruleID, "error", err)
	} else {
		slog.DebugContext(ctx, "updated notification rule message", "runtime", time.Since(start), id, vhost, "ruleID", ruleID)
	}

	return err
}
