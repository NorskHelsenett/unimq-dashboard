package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (dbc *Database) GetNotificationRules(ctx context.Context, id string) ([]models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, id)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "retrieved notification rules", "runtime", time.Since(start), "_id", id, "count", len(notification.Rules))
	return notification.Rules, nil
}

func (dbc *Database) GetNotificationRule(ctx context.Context, vhost string, ruleid string) (*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, vhost)
	if err != nil {
		return nil, err
	}

	for _, rule := range notification.Rules {
		if rule.ID == ruleid {
			slog.InfoContext(ctx, "retrieved notification rule",
				"runtime", time.Since(start),
				"vhost", vhost,
				"rule", rule.Name,
				"_id", rule.ID,
			)
			return &rule, nil
		}
	}
	return nil, fmt.Errorf("notification rule not found")
}

func (dbc *Database) AddNotificationRule(ctx context.Context, vhost string, rule models.AlarmRule) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost}
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
		slog.InfoContext(ctx, "added notification rule",
			"runtime", time.Since(start),
			"vhost", vhost,
			"rule", rule.Name,
		)
	}

	return err
}

func (dbc *Database) DeleteNotificationRule(ctx context.Context, vhost string, id string) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost}
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
			"_id", id,
			"error", err,
		)
		return fmt.Errorf("failed to delete notification rule. %w", err)
	}

	slog.InfoContext(ctx, "deleted notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		"_id", id,
	)
	return nil
}
