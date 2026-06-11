package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (dbc *Database) GetVhost(ctx context.Context, vhost string) (*VhostNotification, error) {
	start := time.Now()
	// cursor, err := dbc.Collections.Notifications.Find(ctx, map[string]any{})
	// if err != nil {
	// 	return nil, err
	// }

	collection := dbc.client.Database(dbc.db).Collection("notifications")

	var notification VhostNotification
	err := collection.FindOne(ctx, map[string]any{"_id": vhost}).Decode(&notification)
	slog.Info("retrieved vhost", "runtime", time.Since(start), "id", vhost)
	return &notification, err
}

func (dbc *Database) CheckVhostExists(ctx context.Context, vhost string) (bool, error) {
	start := time.Now()
	cursor, err := dbc.Collections.Notifications.Find(ctx, map[string]any{"_id": vhost})
	if err != nil {
		return false, err
	}

	defer cursor.Close(ctx)

	var alarms []AlarmEntry
	err = cursor.All(ctx, &alarms)
	if err != nil {
		return false, err
	}

	slog.Info("checked vhost existence", "runtime", time.Since(start), "vhost", vhost, "exists", len(alarms) > 0)
	return len(alarms) > 0, nil
}

func (dbc *Database) EnsureVhostExists(ctx context.Context, name string) error {
	exists, err := dbc.CheckVhostExists(ctx, name)
	if err != nil {
		return err
	}

	if !exists {
		return dbc.AddVhost(ctx, name)
	}

	return nil
}

func (dbc *Database) AddVhost(ctx context.Context, name string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	notification := VhostNotification{
		Name:       name,
		Recipients: []models.Recipient{},
		Rules:      []models.AlarmRule{},
		Notified:   false,
	}

	_, err := collection.InsertOne(ctx, notification)
	slog.Info("added vhost", "runtime", time.Since(start), "name", notification.Name)
	return err
}
