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

	var notification models.VhostNotification
	err := dbc.Collections.Notifications.FindOne(ctx, map[string]any{"_id": vhost}).Decode(&notification)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notification recipient %v. %w", vhost, err)
	}

	for _, recipient := range notification.Recipients {
		if recipient.ID == id {
			slog.InfoContext(ctx, "retrieved notification recipient",
				"runtime", time.Since(start),
				"_id", vhost,
				"recipient", recipient.Name,
				"id", recipient.ID,
			)
			return recipient, nil
		}
	}

	return nil, fmt.Errorf("notification recipient not found. %w", ErrRecipientNotFound)
}

func (dbc *Database) AddNotificationRecipient(ctx context.Context, vhost string, recipient models.Recipient) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost}
	update := map[string]any{
		"$push": map[string]any{
			"recipients": recipient,
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.InfoContext(ctx, "failed to add notification recipient",
			"runtime", time.Since(start),
			"_id", vhost,
			"recipient", recipient.Name,
			"id", recipient.ID,
			"error", err,
		)
	} else {
		slog.InfoContext(ctx, "added notification recipient",
			"runtime", time.Since(start),
			"_id", vhost,
			"recipient", recipient.Name,
			"id", recipient.ID,
		)
	}

	return err
}

func (dbc *Database) DeleteNotificationRecipient(ctx context.Context, vhost string, id string) error {
	start := time.Now()

	filter := map[string]any{"_id": vhost, "recipients.id": id}
	update := map[string]any{
		"$pull": map[string]any{
			"$[].recipients": map[string]any{"id": id},
		},
	}

	_, err := dbc.Collections.Notifications.UpdateOne(ctx, filter, update)
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete notification recipient",
			"runtime", time.Since(start),
			"_id", vhost,
			"id", id,
			"error", err,
		)
	} else {
		slog.InfoContext(ctx, "deleted notification recipient",
			"runtime", time.Since(start),
			"_id", id,
		)
	}

	return err
}
