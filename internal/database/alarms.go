package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/store/notify"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AlarmEntry struct {
	AlarmID string            `bson:"_id"`
	Entries []notify.LogEntry `bson:"entries"`
}

func (dbc *Database) GetAlarmsAll(ctx context.Context) ([]AlarmEntry, error) {

	start := time.Now()

	cursor, err := dbc.Collections.Alarms.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find alarms. %w", err)
	}
	defer cursor.Close(ctx)

	var alarms []AlarmEntry
	err = cursor.All(ctx, &alarms)
	if err != nil {
		return nil, err
	}

	slog.Info("retrieved alarms", "runtime", time.Since(start), "count", len(alarms))

	return alarms, nil
}

func (dbc *Database) GetAlarm(ctx context.Context, id string) (*AlarmEntry, error) {
	start := time.Now()
	var alarm AlarmEntry

	err := dbc.Collections.Alarms.FindOne(ctx, bson.M{"id": id}).Decode(&alarm)
	if err != nil {
		return nil, fmt.Errorf("failed to find alarm. %w", err)
	}

	slog.Info("retrieved alarm", "runtime", time.Since(start), "id", id)

	return &alarm, nil
}

func (dbc *Database) AddAlarm(ctx context.Context, alarm *AlarmEntry) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.InsertOne(ctx, alarm)
	if err != nil {
		return fmt.Errorf("failed to insert alarm. %w", err)
	}
	slog.Info("created alarm", "runtime", time.Since(start), "id", alarm.AlarmID)
	return nil
}

func (dbc *Database) DeleteAlarm(ctx context.Context, alarmID string) error {
	start := time.Now()
	_, err := dbc.Collections.Alarms.DeleteOne(ctx, bson.M{"id": alarmID})
	if err != nil {
		return fmt.Errorf("failed to delete alarm. %w", err)
	}
	slog.Info("deleted alarm", "runtime", time.Since(start), "id", alarmID)
	return nil
}
