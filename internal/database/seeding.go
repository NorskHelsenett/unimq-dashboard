package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	recipientID      = "e45957ef-b817-414e-95e8-c4ea89c4ad3e"
	recipientEmailID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
)

var (
	alarmsSeed = []models.AlarmRule{
		{
			ID:        "090e10a0-4c2c-46e4-8870-9e354232a037",
			Name:      "High queue size",
			Threshold: 15.0,
			Type:      models.AlarmTypeQueueSize,
			Enabled:   true,
			Status:    models.AlarmStatusActive,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "b8f2a1c4-3e7d-4f90-a5b6-1c2d3e4f5a6b",
			Name:      "High channel count",
			Threshold: 10.0,
			Type:      models.AlarmTypeChannels,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "95639405-dfa1-4b3b-b453-bea86abf7bde",
			Name:      "High connection count",
			Threshold: 10.0,
			Type:      models.AlarmTypeConnections,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
	}

	maintenanceEntries = []models.MaintenanceEntry{
		{
			ID:          "e45957ef-b817-414e-95e8-c4ea89c4ad3e",
			Description: "Test maintenance entry 1",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusScheduled,
		},
		{
			ID:          "1fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 2",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusInProgress,
		},
		{
			ID:          "2fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 3",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusDone,
		},
	}
)

func (dbc *Database) Seed(ctx context.Context, vhosts []string) error {
	// Drop stale data so re-seeding always produces a clean state.
	if err := dbc.Collections.Notifications.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop notifications collection: %w", err)
	}
	if err := dbc.Collections.Alarms.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop alarms collection: %w", err)
	}

	for _, vhost := range vhosts {
		err := dbc.seedNotifications(ctx, vhost)
		if err != nil {
			return err
		}

		err = dbc.SeedAlarms(ctx, vhost)
		if err != nil {
			return err
		}
	}

	err := dbc.seedMaintenace(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotifications(ctx context.Context, name string) error {

	err := dbc.EnsureVhostExists(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRecipients(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRules(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRecipients(ctx context.Context, name string) error {

	recipient, err := dbc.GetNotificationRecipient(ctx, name, recipientID)
	if err == nil {
		if recipient == nil {
			return fmt.Errorf("unexpected nil recipient")
		}
		if recipientID != recipient.ID {
			return fmt.Errorf("unexpected recipient ID. expected %s, got %s", recipientID, recipient.ID)
		}
		return nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) && !errors.Is(err, ErrRecipientNotFound) {
		return fmt.Errorf("failed to check existing recipient entry. %w", err)
	}

	slog.InfoContext(ctx, "seeding notification recipient", "vhost", name, "recipientID", recipientID, "existing", recipient)

	err = dbc.AddNotificationRecipient(ctx, name, &models.Recipient{
		ID:   recipientID,
		Name: "Matias",
		URL:  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
		Type: models.RecipientTypeWebhook,
	})
	if err != nil {
		return err
	}

	err = dbc.AddNotificationRecipient(ctx, name, &models.Recipient{
		ID:    recipientEmailID,
		Name:  "Matias",
		Email: "matias.nordmann@example.no",
		Type:  models.RecipientTypeEmail,
	})
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRules(ctx context.Context, vhost string) error {

	rules, err := dbc.GetNotificationRules(ctx, vhost)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("failed to get existing notification rules. %w", err)
		}
	}

	missingRules := []models.AlarmRule{}
	for _, rule := range alarmsSeed {
		found := false
		if rules != nil {
			for _, existingRule := range rules {
				if rule.ID == existingRule.ID {
					found = true
					break
				}
			}
		}
		if !found {
			missingRules = append(missingRules, rule)
		}

	}

	if len(missingRules) == 0 {
		return nil
	}

	for _, missingRule := range missingRules {
		err = dbc.AddNotificationRule(ctx, vhost, &missingRule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dbc *Database) seedMaintenace(ctx context.Context) error {

	for _, MaintenanceEntry := range maintenanceEntries {
		existing, err := dbc.GetMaintenanceEntry(ctx, MaintenanceEntry.ID)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("failed to check existing maintenance entry. %w", err)
			}
		} else {
			if existing == nil {
				return fmt.Errorf("unexpected nil maintenance entry")
			}
			if existing.ID == MaintenanceEntry.ID {
				continue
			}
			return fmt.Errorf("unexpected maintenance entry ID. expected %s, got %s", "test-maintenance", existing.ID)
		}

		err = dbc.AddMaintenanceEntry(ctx, &MaintenanceEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbc *Database) SeedMaintenanceEntry(ctx context.Context, entry *models.MaintenanceEntry) error {
	existing, err := dbc.GetMaintenanceEntry(ctx, entry.ID)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("failed to check existing maintenance entry. %w", err)
		}
	} else {
		if existing == nil {
			return fmt.Errorf("unexpected nil maintenance entry")
		}
		if existing.ID == entry.ID {
			return nil
		}
		return fmt.Errorf("unexpected maintenance entry ID. expected %s, got %s", entry.ID, existing.ID)
	}

	err = dbc.AddMaintenanceEntry(ctx, entry)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) SeedMaintenanceLogs(ctx context.Context) error {

	for _, MaintenanceEntry := range maintenanceEntries {

		logEntries := []models.MaintenanceEditLog{}
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 1", time.Now(), time.Now().Add(2*time.Hour), "Initial creation", "user1"))
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 2", time.Now(), time.Now().Add(2*time.Hour), "Updated description", "user2"))
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 3", time.Now(), time.Now().Add(2*time.Hour), "Updated start and end times", "user3"))

		missingEntries := []models.MaintenanceEditLog{}
		existing, err := dbc.GetMaintenanceEditLogs(ctx, MaintenanceEntry.ID)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("failed to check existing maintenance edit logs. %w", err)
			}
		} else {
			if len(existing) > 0 {
				for _, log := range logEntries {
					found := false
					for _, existingLog := range existing {
						if log.MaintenanceID == existingLog.MaintenanceID {
							found = true
							break
						}
					}
					if !found {
						missingEntries = append(missingEntries, log)
					}

				}
			} else {
				missingEntries = logEntries
			}
		}

		if len(missingEntries) > 0 {
			for _, log := range missingEntries {
				err = dbc.AddMaintenanceEditLog(ctx, &log)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func (dbc *Database) SeedAlarms(ctx context.Context, name string) error {
	for _, rule := range alarmsSeed {

		alarm, err := dbc.GetAlarm(ctx, rule.ID)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("failed to check existing alarm entry. %w", err)
			}
		} else {
			if alarm == nil {
				return fmt.Errorf("unexpected nil alarm entry")
			}

			if alarm.AlarmID != rule.ID {
				return fmt.Errorf("unexpected alarm entry name. expected %s, got %s", "name", alarm.AlarmID)
			}
			return nil
		}

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

		err = dbc.AddAlarm(ctx, &entries)
		if err != nil {
			return err
		}
	}

	// This is mostly to test the InsertAlarmEntries function, which is used to insert log entries for a specific vhost.
	alarms := make([]*models.LogEntry, 0, 2)
	entry := models.NewLogEntry(models.LogEventFired, new(5.0), 1.337, models.AlarmTypeQueueMessages)
	alarms = append(alarms, &entry)
	entry = models.NewLogEntry(models.LogEventResolved, new(5.0), 5.318008, models.AlarmTypeQueueMessages)
	alarms = append(alarms, &entry)

	err := dbc.InsertAlarmEntries(ctx, name, alarms)
	if err != nil {
		return err
	}

	return nil
}
