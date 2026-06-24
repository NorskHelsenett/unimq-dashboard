package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/bodycloserhelper"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetVhost(ctx context.Context, vhost string) (*models.VhostNotification, error) {
	start := time.Now()

	var notification models.VhostNotification
	err := dbc.Collections.Notifications.FindOne(ctx, bson.M{"_id": vhost}).Decode(&notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve vhost", "runtime", time.Since(start), "_id", vhost, "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "retrieved vhost", "runtime", time.Since(start), "_id", vhost)
	return &notification, err
}

func (dbc *Database) CheckVhostExists(ctx context.Context, vhost string) (bool, error) {
	start := time.Now()
	cursor, err := dbc.Collections.Notifications.Find(ctx, map[string]any{"_id": vhost})
	if err != nil {
		return false, err
	}

	defer bodycloserhelper.BodyCloseResponse(cursor.Close(ctx))

	var alarms []models.AlarmEntry
	err = cursor.All(ctx, &alarms)
	if err != nil {
		slog.ErrorContext(ctx, "checked vhost existence", "runtime", time.Since(start), "_id", vhost, "exists", len(alarms) > 0, "error", err)
		return false, err
	}

	slog.InfoContext(ctx, "checked vhost existence", "runtime", time.Since(start), "_id", vhost, "exists", len(alarms) > 0)
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

	notification := models.NewVhostNotification(name)
	_, err := dbc.Collections.Notifications.InsertOne(ctx, notification)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add vhost", "runtime", time.Since(start), "_id", name, "error", err)
	} else {
		slog.InfoContext(ctx, "added vhost", "runtime", time.Since(start), "_id", notification.Name)
	}

	return err
}
