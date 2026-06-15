package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

var (
	ErrRecipientNotFound = fmt.Errorf("notification recipient not found")
)

func (dbc *Database) GetNotificationRecipient(ctx context.Context, vhost string, id string) (*models.Recipient, error) {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	var notification VhostNotification
	err := collection.FindOne(ctx, map[string]any{"_id": vhost}).Decode(&notification)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notification recipient %v. %w", vhost, err)
	}

	for _, recipient := range notification.Recipients {
		if recipient.ID == id {
			slog.Info("retrieved notification recipient",
				"runtime", time.Since(start),
				"vhost", vhost,
				"recipient", recipient.Name,
				"_id", recipient.ID,
			)
			return &recipient, nil
		}
	}

	return nil, fmt.Errorf("notification recipient not found. %w", ErrRecipientNotFound)
}

func (dbc *Database) AddNotificationRecipient(ctx context.Context, vhost string, recipient models.Recipient) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"_id": vhost},
		map[string]any{
			"$push": map[string]any{
				"recipients": recipient,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to add notification recipient. %w", err)
	}

	slog.Info("added notification recipient",
		"runtime", time.Since(start),
		"vhost", vhost,
		"recipient", recipient.Name,
		"_id", recipient.ID,
	)
	return nil
}

func (dbc *Database) DeleteNotificationRecipient(ctx context.Context, vhost string, id string) error {
	start := time.Now()
	collection := dbc.client.Database(dbc.db).Collection("notifications")

	_, err := collection.UpdateOne(
		ctx,
		map[string]any{"_id": vhost, "recipients.id": id},
		map[string]any{
			"$pull": map[string]any{
				"$[].recipients": map[string]any{"id": id},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete notification recipient. %w", err)
	}

	slog.Info("deleted notification recipient",
		"runtime", time.Since(start),
		"_id", id,
	)
	return nil
}
