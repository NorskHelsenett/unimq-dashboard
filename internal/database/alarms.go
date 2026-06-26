package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (dbc *Database) GetAlarmsAll(ctx context.Context) ([]models.AlarmEntry, error) {

	start := time.Now()

	cursor, err := dbc.Collections.Alarms.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find alarms. %w", err)
	}

	defer func() {
		err := cursor.Close(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to close cursor", "runtime", time.Since(start), "error", err)
		}
	}()

	var alarms []models.AlarmEntry
	err = cursor.All(ctx, &alarms)
	if err != nil {
		slog.ErrorContext(ctx, "failed to decode alarms", "runtime", time.Since(start), "error", err)
		return nil, err
	}

	slog.DebugContext(ctx, "retrieved alarms", "runtime", time.Since(start), "count", len(alarms))

	return alarms, nil
}

func (dbc *Database) GetAlarm(ctx context.Context, id string) (*models.AlarmEntry, error) {
	start := time.Now()
	var alarm models.AlarmEntry

	err := dbc.Collections.Alarms.FindOne(ctx, bson.M{"_id": id}).Decode(&alarm)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find alarm", "runtime", time.Since(start), "_id", id, "error", err)
		return nil, fmt.Errorf("failed to find alarm. %w", err)
	}

	slog.DebugContext(ctx, "retrieved alarm", "runtime", time.Since(start), "_id", id)

	return &alarm, nil
}

func (dbc *Database) AddAlarm(ctx context.Context, alarm *models.AlarmEntry) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.InsertOne(ctx, alarm)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create alarm", "runtime", time.Since(start), "_id", alarm.AlarmID, "error", err)
	} else {
		slog.DebugContext(ctx, "created alarm", "runtime", time.Since(start), "_id", alarm.AlarmID)

	}
	return err
}

func (dbc *Database) DeleteAlarm(ctx context.Context, alarmID string) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.DeleteOne(ctx, bson.M{"_id": alarmID})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete alarm", "runtime", time.Since(start), "_id", alarmID, "error", err)
	} else {
		slog.DebugContext(ctx, "deleted alarm", "runtime", time.Since(start), "_id", alarmID)
	}
	return err
}
