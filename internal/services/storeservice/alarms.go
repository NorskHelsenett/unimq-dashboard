package storeservice

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (rc *StoreService) SeedAlarms(ctx context.Context, alarms []*models.AlarmRule) error {

	// nolint:gocritic // ignoring rangeValCopy as this is only used for seeding and not performance critical.
	for _, rule := range alarms {

		entries := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{},
		}

		val := 5.0
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, &val, 3.4, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventResolved, &val, 42.0, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, &val, 0.67, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventResolved, &val, 69.0, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, new(47.0), 10.0, rule.Type))

		err := rc.AlarmHandler().Create(ctx, []*models.AlarmEntry{&entries})
		if err != nil {
			return err
		}
	}

	return nil
}
