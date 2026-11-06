package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AlarmHandler struct {
	Collection *mongo.Collection
	*CollectorHolder[models.AlarmEntry]
}

func NewAlarmHandler(collection *mongo.Collection) *AlarmHandler {
	return &AlarmHandler{
		Collection:      collection,
		CollectorHolder: NewCollectorHolder[models.AlarmEntry](collection),
	}
}

// interface guard clause.
var _ database.AlarmRepository = (*AlarmHandler)(nil)

var (
	ErrAlarmNotFound = fmt.Errorf("alarm not found")
)

func (dbc *Database) InsertAlarmEntries(ctx context.Context, alarmID string, logEntries []models.LogEntry) error {
	start := time.Now()
	filter := bson.M{id: alarmID}
	update := bson.M{"$push": bson.M{"entries": bson.M{"$each": logEntries}}}
	results, err := dbc.Collections.Alarms.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		slog.ErrorContext(ctx, "failed to update alarm",
			"runtime", time.Since(start),
			id, alarmID,
			"error", err,
		)
		return fmt.Errorf("failed to update alarm. %w", err)
	}

	if results.UpsertedCount == 0 && results.MatchedCount == 0 {
		slog.ErrorContext(ctx, "failed to update alarm",
			"runtime", time.Since(start),
			id, alarmID,
			"error", "no documents updated",
		)
		return fmt.Errorf("failed to update alarm. no documents updated. %w", ErrAlarmNotFound)
	}

	slog.DebugContext(ctx, "updated alarm", "runtime", time.Since(start), id, alarmID)
	return nil
}

func (ah *AlarmHandler) Append(ctx context.Context, alarmID string, logEntry models.LogEntry) error {
	return fmt.Errorf("not implemented")
}
