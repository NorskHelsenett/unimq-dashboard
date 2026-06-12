package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type VhostNotification struct {
	Name       string             `bson:"_id"`
	Recipients []models.Recipient `bson:"recipients"`
	Rules      []models.AlarmRule `bson:"rules"`
	Notified   bool               `bson:"notified"`
}

func (vn *VhostNotification) WebhookURLs() []string {
	urls := make([]string, 0, len(vn.Recipients))
	for _, r := range vn.Recipients {
		if r.Type == models.RuleTypeWebhook {
			urls = append(urls, r.URL)
		}
	}
	return urls
}

// type VhostNotification struct {
// 	Name     string                        `bson:"name"`
// 	Vhosts   map[string]models.VhostConfig `bson:"vhosts"`
// 	Notified bool                          `bson:"notified"`
// }

func (dbc *Database) GetNotificationsAll(ctx context.Context) ([]VhostNotification, error) {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	cursor, err := collection.Find(ctx, map[string]any{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []VhostNotification
	err = cursor.All(ctx, &notifications)
	if err != nil {
		return nil, err
	}

	slog.Info("retrieved notifications", "runtime", time.Since(start), "count", len(notifications))
	return notifications, nil
}

// Probably unnecessary as vhost functions already cover this.
func (dbc *Database) GetNotification(ctx context.Context, vhost string) (VhostNotification, error) {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	var notification VhostNotification
	err := collection.FindOne(ctx, map[string]any{"vhost": vhost}).Decode(&notification)
	slog.Info("retrieved notification", "runtime", time.Since(start), "id", vhost)
	return notification, err
}

func (dbc *Database) AddNotification(ctx context.Context, notification VhostNotification) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.InsertOne(ctx, notification)
	slog.Info("added notification", "runtime", time.Since(start), "name", notification.Name)
	return err
}

func (dbc *Database) UpdateNotification(ctx context.Context, name string, notification VhostNotification) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"name": name},
		map[string]any{
			"$set": notification,
		},
	)
	slog.Info("updated notification", "runtime", time.Since(start), "name", notification.Name)
	return err
}

func (dbc *Database) DeleteNotification(ctx context.Context, path string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.DeleteOne(ctx, map[string]any{"path": path})
	slog.Info("deleted notification", "runtime", time.Since(start), "path", path)
	return err
}

func (dbc *Database) UpdateNotificationRuleThreshold(ctx context.Context, vhost, ruleID string, threshold float64) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"name": vhost, "rules.id": ruleID},
		map[string]any{
			"$set": map[string]any{
				"rules.$.threshold": threshold,
			},
		},
	)
	slog.Info("updated notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID)
	return err
}

func (dbc *Database) UpdateNotificationRuleMessage(ctx context.Context, vhost, ruleID string, message string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"name": vhost, "rules.id": ruleID},
		map[string]any{
			"$set": map[string]any{
				"rules.$.message": message,
			},
		},
	)
	slog.Info("updated notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID)
	return err
}

func (dbc *Database) UpdateNotificationRule(ctx context.Context, vhost, name string, status string, value float64, notified bool) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	// ID        string     `json:"id" bson:"_id"`
	// Name      string     `json:"name" bson:"name"`
	// Type      string     `json:"type" bson:"type"`
	// QueueName string     `json:"queue_name,omitempty" bson:"queueName"`
	// Threshold float64    `json:"threshold,omitempty" bson:"threshold"`
	// Message   string     `json:"message" bson:"message"`
	// Enabled   bool       `json:"enabled" bson:"enabled"`
	// Status    string     `json:"status" bson:"status"`
	// LastFired *time.Time `json:"last_fired,omitempty" bson:"lastFired"`
	// LastValue *float64   `json:"last_value,omitempty" bson:"lastValue"`
	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"_id": vhost, "rules.name": name},
		map[string]any{
			"$set": map[string]any{
				"rules.$[rule].status":    status,
				"rules.$[rule].lastValue": value,
				"notified":                notified,
			},
		},
	)
	slog.Info("updated notification status", "runtime", time.Since(start), "name", name, "notified", notified)
	return err
}

func (dbc *Database) ToggleNotificationRule(ctx context.Context, vhost, ruleID string, enabled bool) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"name": vhost, "rules.id": ruleID},
		map[string]any{
			"$set": map[string]any{
				"rules.$.enabled": enabled,
			},
		},
	)
	slog.Info("toggled notification rule", "runtime", time.Since(start), "vhost", vhost, "ruleID", ruleID, "enabled", enabled)
	return err
}
