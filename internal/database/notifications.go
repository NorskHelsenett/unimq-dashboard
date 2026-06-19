package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetNotificationsAll(ctx context.Context) ([]models.VhostNotification, error) {
	start := time.Now()

	cursor, err := dbc.Collections.Notifications.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.VhostNotification
	err = cursor.All(ctx, &notifications)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "retrieved notifications", "runtime", time.Since(start), "count", len(notifications))
	return notifications, nil
}

// Probably unnecessary as vhost functions already cover this.
func (dbc *Database) GetNotification(ctx context.Context, vhost string) (*models.VhostNotification, error) {
	start := time.Now()

	var notification models.VhostNotification
	err := dbc.Collections.Notifications.FindOne(ctx, bson.M{"_id": vhost}).Decode(&notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve notification", "runtime", time.Since(start), "_id", vhost, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "retrieved notification", "runtime", time.Since(start), "_id", vhost)
	return &notification, nil
}

func (dbc *Database) AddNotification(ctx context.Context, notification models.VhostNotification) error {
	start := time.Now()

	_, err := dbc.Collections.Notifications.InsertOne(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add notification", "runtime", time.Since(start), "name", notification.Name, "error", err)
		return err
	}

	slog.InfoContext(ctx, "added notification", "runtime", time.Since(start), "name", notification.Name)
	return err
}

func (dbc *Database) UpdateNotification(ctx context.Context, name string, notification models.VhostNotification) error {
	start := time.Now()

	filter := map[string]any{"_id": name}
	update := map[string]any{
		"$set": notification,
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification", "runtime", time.Since(start), "_id", name, "error", err)
	} else {
		slog.InfoContext(ctx, "updated notification", "runtime", time.Since(start), "_id", notification.Name)
	}

	return err
}

func (dbc *Database) DeleteNotification(ctx context.Context, id string) error {
	start := time.Now()

	_, err := dbc.Collections.Notifications.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete notification", "runtime", time.Since(start), "_id", id, "error", err)
	} else {
		slog.InfoContext(ctx, "deleted notification", "runtime", time.Since(start), "_id", id)
	}

	return err
}

func (dbc *Database) UpdateNotificationRuleThreshold(ctx context.Context, vhost, ruleID string, threshold float64) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost, "rules.id": ruleID}
	update := map[string]any{
		"$set": map[string]any{
			"rules.$.threshold": threshold,
		},
	}
	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID, "error", err)
	} else {
		slog.InfoContext(ctx, "updated notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID)
	}

	return err
}

func (dbc *Database) UpdateNotificationRuleMessage(ctx context.Context, vhost, ruleID string, message string) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost, "rules.id": ruleID}
	update := map[string]any{
		"$set": map[string]any{
			"rules.$.message": message,
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule message", "runtime", time.Since(start), "_id", vhost, "ruleID", ruleID, "error", err)
	} else {
		slog.InfoContext(ctx, "updated notification rule message", "runtime", time.Since(start), "_id", vhost, "ruleID", ruleID)
	}

	return err
}

func (dbc *Database) UpdateNotificationRule(ctx context.Context, vhost, name string, status string, value float64, notified bool) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost, "rules.name": name}
	update := map[string]any{
		"$set": map[string]any{
			"rules.$.status":    status,
			"rules.$.lastValue": value,
			"notified":          notified,
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification rule status", "runtime", time.Since(start), "_id", vhost, "rule", name, "error", err)
		return err
	} else {
		slog.InfoContext(ctx, "updated notification rule status", "runtime", time.Since(start), "_id", vhost, "name", name, "notified", notified)
	}

	return err
}

func (dbc *Database) ToggleNotificationRule(ctx context.Context, vhost, ruleID string, enabled bool) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost, "rules.id": ruleID}
	update := map[string]any{
		"$set": map[string]any{
			"rules.$.enabled": enabled,
		},
	}
	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to toggle notification rule", "runtime", time.Since(start), "_id", vhost, "ruleID", ruleID, "enabled", enabled, "error", err)
		return err
	}

	slog.InfoContext(ctx, "toggled notification rule", "runtime", time.Since(start), "_id", vhost, "ruleID", ruleID, "enabled", enabled)
	return err
}
