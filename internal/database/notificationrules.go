package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (dbc *Database) GetNotificationRules(ctx context.Context, vhost string) ([]models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, vhost)
	if err != nil {
		return nil, err
	}

	slog.Info("retrieved notification rules", "runtime", time.Since(start), "vhost", vhost, "count", len(notification.Rules))
	return notification.Rules, nil
}

func (dbc *Database) GetNotificationRule(ctx context.Context, vhost string, id string) (*models.AlarmRule, error) {
	start := time.Now()
	notification, err := dbc.GetVhost(ctx, vhost)
	if err != nil {
		return nil, err
	}

	for _, rule := range notification.Rules {
		if rule.ID == id {
			slog.Info("retrieved notification rule",
				"runtime", time.Since(start),
				"vhost", vhost,
				"rule", rule.Name,
				"id", rule.ID,
			)
			return &rule, nil
		}
	}
	return nil, fmt.Errorf("notification rule not found")
}

func (dbc *Database) AddNotificationRule(ctx context.Context, vhost string, rule models.AlarmRule) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"_id": vhost},
		map[string]any{
			"$push": map[string]any{
				"rules": rule,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to add notification rule. %w", err)
	}

	slog.Info("added notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		"rule", rule.Name,
	)
	return nil
}

func (dbc *Database) DeleteNotificationRule(ctx context.Context, vhost string, id string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"_id": vhost},
		map[string]any{
			"$pull": map[string]any{
				"rules": map[string]any{"id": id},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete notification rule. %w", err)
	}

	slog.Info("deleted notification rule",
		"runtime", time.Since(start),
		"vhost", vhost,
		"id", id,
	)
	return nil
}
