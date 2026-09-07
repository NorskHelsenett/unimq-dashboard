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

func (dbc *Database) GetAlarm(ctx context.Context, alarmID string) (*models.AlarmEntry, error) {
	start := time.Now()
	var alarm models.AlarmEntry

	filter := bson.M{id: alarmID}
	err := dbc.Collections.Alarms.FindOne(ctx, filter).Decode(&alarm)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find alarm", "runtime", time.Since(start), "_id", alarmID, "error", err)
		return nil, fmt.Errorf("failed to find alarm. %w", err)
	}

	slog.DebugContext(ctx, "retrieved alarm", "runtime", time.Since(start), "_id", alarmID)

	return &alarm, nil
}

func (dbc *Database) AddAlarm(ctx context.Context, alarm *models.AlarmEntry) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.InsertOne(ctx, alarm)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create alarm", "runtime", time.Since(start), id, alarm.AlarmID, "error", err)
	} else {
		slog.DebugContext(ctx, "created alarm", "runtime", time.Since(start), id, alarm.AlarmID)

	}
	return err
}

func (dbc *Database) InsertAlarmEntries(ctx context.Context, alarmID string, logEntries []*models.LogEntry) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.UpdateOne(ctx, bson.M{id: alarmID}, bson.M{"$push": bson.M{"entries": bson.M{"$each": logEntries}}})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update alarm", "runtime", time.Since(start), id, alarmID, "error", err)
	} else {
		slog.DebugContext(ctx, "updated alarm", "runtime", time.Since(start), id, alarmID)
	}
	return err
}

func (dbc *Database) DeleteAlarm(ctx context.Context, alarmID string) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.DeleteOne(ctx, bson.M{id: alarmID})
	if err != nil {
		slog.ErrorContext(ctx, "failed to delete alarm", "runtime", time.Since(start), id, alarmID, "error", err)
	} else {
		slog.DebugContext(ctx, "deleted alarm", "runtime", time.Since(start), id, alarmID)
	}
	return err
}
