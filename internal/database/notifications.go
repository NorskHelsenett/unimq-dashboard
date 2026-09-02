package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (dbc *Database) GetNotificationsAll(ctx context.Context) ([]models.VhostNotification, error) {
	start := time.Now()

	cursor, err := dbc.Collections.Notifications.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to close cursor", "runtime", time.Since(start), "error", err)
		}
	}()

	var notifications []models.VhostNotification
	err = cursor.All(ctx, &notifications)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved notifications", "runtime", time.Since(start), "count", len(notifications))
	return notifications, nil
}

// Probably unnecessary as vhost functions already cover this.
func (dbc *Database) GetNotification(ctx context.Context, vhost string) (*models.VhostNotification, error) {
	start := time.Now()

	var notification models.VhostNotification
	err := dbc.Collections.Notifications.FindOne(ctx, bson.M{id: vhost}).Decode(&notification)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.DebugContext(ctx, "notification not found", "runtime", time.Since(start), id, vhost)
			return nil, fmt.Errorf("notification not found for vhost %s. %w", vhost, err)
		}
		slog.ErrorContext(ctx, "failed to retrieve notification", "runtime", time.Since(start), id, vhost, "error", err)
		return nil, fmt.Errorf("failed to retrieve notification for vhost %s. %w", vhost, err)
	}

	slog.DebugContext(ctx, "retrieved notification", "runtime", time.Since(start), id, vhost)
	return &notification, nil
}

func (dbc *Database) AddNotification(ctx context.Context, notification models.VhostNotification) error {
	start := time.Now()

	_, err := dbc.Collections.Notifications.InsertOne(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add notification", "runtime", time.Since(start), "name", notification.Name, "error", err)
		return err
	}

	slog.DebugContext(ctx, "added notification", "runtime", time.Since(start), "name", notification.Name)
	return err
}

func (dbc *Database) UpdateNotification(ctx context.Context, name string, notification models.VhostNotification) error {
	start := time.Now()

	filter := map[string]any{id: name}
	update := map[string]any{
		set: notification,
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update notification", "runtime", time.Since(start), id, name, "error", err)
	} else {
		slog.DebugContext(ctx, "updated notification", "runtime", time.Since(start), id, notification.Name)
	}

	return err
}

func (dbc *Database) DeleteNotification(ctx context.Context, notificationID string) error {
	start := time.Now()

	status, err := dbc.Collections.Notifications.DeleteOne(ctx, bson.M{id: notificationID})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete notification", "runtime", time.Since(start), id, notificationID, "error", err)
		return err
	}

	if status.DeletedCount == 0 {
		slog.ErrorContext(ctx, "no notification found to delete", "runtime", time.Since(start), id, notificationID)
		return fmt.Errorf("notification not found for vhost %s. %w", id, mongo.ErrNoDocuments)
	}

	slog.DebugContext(ctx, "deleted notification", "runtime", time.Since(start), id, notificationID)

	return nil
}
