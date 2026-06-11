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
