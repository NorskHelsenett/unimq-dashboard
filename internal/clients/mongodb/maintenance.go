package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrMaintenanceNotFound = fmt.Errorf("maintenance entry not found")
)

type MaintenaceHandler struct {
	Collection mongo.Collection
	CollectorHolder[models.MaintenanceEntry]
}

func NewMaintenanceHandler(collection mongo.Collection) *MaintenaceHandler {
	return &MaintenaceHandler{
		Collection:      collection,
		CollectorHolder: CollectorHolder[models.MaintenanceEntry]{},
	}
}

// interface guard clause.
var _ database.MaintenaceRepository = (*MaintenaceHandler)(nil)

func (mh *MaintenaceHandler) Advance(ctx context.Context) error {
	start := time.Now()
	now := start.UTC()

	filter := []database.Filter{
		database.NewFilter("status", database.FilterOpEqual, models.MaintenanceStatusScheduled.String()),
		database.NewFilter("start", database.FilterOpLessEqual, now),
		database.NewFilter("end", database.FilterOpGreater, now),
	}

	update := map[string]any{"status": models.MaintenanceStatusInProgress}

	// scheduled → in_progress when start has passed but end has not
	err := mh.Patch(ctx, update, filter...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to advance maintenance statuses to in_progress",
			"runtime", time.Since(start),
			"error", err,
		)
		return err
	}

	filter = []database.Filter{
		database.NewFilter("status", database.FilterOpIn, []string{
			models.MaintenanceStatusScheduled.String(),
			models.MaintenanceStatusInProgress.String(),
		}),
		database.NewFilter("end", database.FilterOpLessEqual, now),
	}
	update = map[string]any{"status": models.MaintenanceStatusDone}

	// scheduled or in_progress → done when end has passed
	err = mh.Patch(ctx, update, filter...)
	if err != nil {
		slog.ErrorContext(ctx, "failed to advance maintenance statuses to done",
			"runtime", time.Since(start),
			"error", err,
		)
		return err
	}

	return nil
}
